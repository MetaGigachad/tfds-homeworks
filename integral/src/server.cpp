#include "server.hpp"

#include <arpa/inet.h>
#include <sys/epoll.h>
#include <sys/socket.h>
#include <unistd.h>

#include <cstring>
#include <iostream>
#include <string>

#include "integral.hpp"
#include "logger.hpp"
#include "msg.hpp"

namespace integral {

Server::Server(const Config& config)
    : Config_(config), EpollFd_(-1), Result_(0) {
  EpollFd_ = epoll_create1(0);
  if (EpollFd_ == -1) {
    throw std::runtime_error("Error creating epoll instance");
  }
}

void Server::Start() {
  Log.Info("Starting server");

  auto udpSock = MakeUdpSocket();
  auto broadcastAddr = MakeBroadcastAddr();
  BroadcastHello(udpSock, broadcastAddr);
  EnqueueCalculation();
  EpollEventLoop(udpSock);
}

void Server::EnqueueCalculation() {
  for (auto task : Split(ComputeIntervalTask({0, 1}), std::stol(Config_.at("taskCount")))) {
    TaskQueue_.emplace_back(std::move(task));
  }
}

int Server::MakeUdpSocket() {
  int sock = socket(AF_INET, SOCK_DGRAM, 0);
  if (sock < 0) {
    throw std::runtime_error("Socket creation failed");
  }

  int broadcastPermission = 1;
  if (setsockopt(sock, SOL_SOCKET, SO_BROADCAST, &broadcastPermission,
                 sizeof(broadcastPermission)) < 0) {
    throw std::runtime_error("Setting broadcast permission failed");
  }

  return sock;
}

sockaddr_in Server::MakeBroadcastAddr() {
  struct sockaddr_in broadcastAddr;
  std::memset(&broadcastAddr, 0, sizeof(broadcastAddr));
  broadcastAddr.sin_family = AF_INET;
  broadcastAddr.sin_port = htons(std::stoul(Config_.at("broadcastPort")));

  if (inet_pton(AF_INET, Config_.at("broadcastIp").c_str(),
                &broadcastAddr.sin_addr) <= 0) {
    throw std::runtime_error("Invalid broadcast address");
  }
  return broadcastAddr;
}

void Server::BroadcastHello(int sock, sockaddr_in broadcastAddr) {
  Log.Debug("Broadcasting Hello");

  const char message[] = {static_cast<char>(MsgCode::Hello)};
  if (sendto(sock, message, sizeof(message), 0,
             reinterpret_cast<sockaddr*>(&broadcastAddr),
             sizeof(broadcastAddr)) < 0) {
    throw std::runtime_error("Sending broadcast failed");
  }
}

void Server::EpollEventLoop(int udpSock) {
  struct epoll_event ev;
  ev.events = EPOLLIN;
  ev.data.fd = udpSock;
  if (epoll_ctl(EpollFd_, EPOLL_CTL_ADD, udpSock, &ev) == -1) {
    throw std::runtime_error("Error adding UDP socket to epoll");
  }

  const auto pingInterval = 5;
  time_t lastPingTime = time(NULL);

  const auto maxEvents = 10;
  epoll_event events[maxEvents];

  while (true) {
    try {
      int nEvents = epoll_wait(EpollFd_, events, maxEvents, 50);
      if (nEvents == -1) {
        throw std::runtime_error("epoll_wait failed");
      }

      for (int i = 0; i < nEvents; ++i) {
        if (events[i].data.fd == udpSock) {
          HandleUdpMessage(events[i].data.fd);
        } else {
          HandleTcpMessage(events[i].data.fd);
        }
      }

      if (difftime(time(NULL), lastPingTime) >= pingInterval) {
        if (Workers_.empty()) {
          auto broadcastAddr = MakeBroadcastAddr();
          BroadcastHello(udpSock, broadcastAddr);
        } else {
          PingWorkers();
        }
        lastPingTime = time(NULL);
      }

      if (!Workers_.empty()) {
        AssignCalc();
      }

      if (TaskQueue_.empty() && WorkerToTask_.empty()) {
        break;
      }
    } catch (const std::runtime_error& err) {
      Log.Error("Error occured in epoll event loop: {}", err.what());
    }
  }

  Log.Info("Result={}", Result_);
}

void Server::HandleUdpMessage(int sock) {
  sockaddr_in recvAddr;
  socklen_t addrLen = sizeof(recvAddr);
  char buffer[1024];

  int n = recvfrom(sock, buffer, sizeof(buffer) - 1, 0,
                   reinterpret_cast<sockaddr*>(&recvAddr), &addrLen);

  buffer[n] = '\0';
  if (n != 1 || buffer[0] != static_cast<char>(MsgCode::HelloAck)) {
    throw std::runtime_error("Invalid udp message");
  }

  ConnectToWorker(recvAddr);
}

void Server::ConnectToWorker(sockaddr_in& recvAddr) {
  int tcpSock = socket(AF_INET, SOCK_STREAM, 0);
  if (tcpSock < 0) {
    std::cerr << "Error creating TCP socket\n";
    return;
  }

  auto workerAddr = recvAddr;
  workerAddr.sin_port = htons(std::stol(Config_.at("workerTcpPort")));
  char ipAddr[16];
  inet_ntop(AF_INET, &workerAddr.sin_addr, ipAddr, sizeof(ipAddr));
  auto key = std::format("{}:{}", ipAddr, workerAddr.sin_port);

  if (WorkerAddrs_.contains(key)) {
    std::cerr << "Worker already exists \n";
    return;
  }

  if (connect(tcpSock, reinterpret_cast<sockaddr*>(&workerAddr),
              sizeof(workerAddr)) < 0) {
    close(tcpSock);
    throw std::runtime_error("Error connecting to worker");
  }

  struct epoll_event ev;
  ev.events = EPOLLIN;
  ev.data.fd = tcpSock;
  if (epoll_ctl(EpollFd_, EPOLL_CTL_ADD, tcpSock, &ev) == -1) {
    throw std::runtime_error("Error adding TCP socket to epoll");
  }

  WorkerAddrs_.insert(key);
  WorkerConnection worker = {true, true, key};
  Workers_[tcpSock] = worker;
  Log.Debug("Connected to worker id={}", key);
}

void Server::HandleTcpMessage(int tcpSock) {
  char msgCode[1];
  ssize_t bytes_received = recv(tcpSock, &msgCode, sizeof(msgCode), 0);
  if (bytes_received == 0) {
    auto worker = Workers_[tcpSock];
    WorkerAddrs_.erase(worker.Key);
    Workers_.erase(tcpSock);
    if (epoll_ctl(EpollFd_, EPOLL_CTL_DEL, tcpSock, NULL) == -1) {
      throw std::runtime_error(
          "Error while removing worker's tcpSocket from epoll");
    }
    Log.Debug("Disconnected from worker id={}", worker.Key);
    if (WorkerToTask_.contains(tcpSock)) {
      TaskQueue_.push_front(WorkerToTask_[tcpSock]);
      WorkerToTask_.erase(tcpSock);
    }
    return;
  } else if (bytes_received != 1) {
    throw std::runtime_error("Error reading message code");
  }

  if (msgCode[0] == static_cast<char>(MsgCode::PingAck)) {
    Log.Debug("Got PingAck from worker id={}", Workers_[tcpSock].Key);
    Workers_[tcpSock].RecievedLastPingAck = true;
    Workers_[tcpSock].Alive = true;
  } else if (msgCode[0] == static_cast<char>(MsgCode::CalcResult)) {
    CalcResult result;
    if (recv(tcpSock, &result, sizeof(result), 0) != sizeof(result)) {
      throw std::runtime_error("Error reading result from worker");
    }
    if (WorkerToTask_.contains(tcpSock) &&
        WorkerToTask_[tcpSock].Interval == result.Range) {
      WorkerToTask_.erase(tcpSock);
    }
    if (!CalculatedIntervals_.contains(result.Range)) {
      CalculatedIntervals_.insert(result.Range);
      Result_ += result.Result;
      Log.Info("Calculation success on range=({})--({}) result={}",
                result.Range.first, result.Range.second, result.Result);
    }
  }
}

void Server::PingWorkers() {
  Log.Debug("Pinging workers");

  for (auto& [sock, worker] : Workers_) {
    const char msg[] = {static_cast<char>(MsgCode::Ping)};
    if (send(sock, msg, sizeof(msg), 0) < 0) {
      throw std::runtime_error(std::format("Error while sending Ping"));
    }
    if (!worker.RecievedLastPingAck) {
      worker.Alive = false;
    } else {
      worker.RecievedLastPingAck = false;
    }
  }
}

void Server::AssignCalc() {
  while (true) {
    ComputeIntervalTask task;
    if (!TaskQueue_.empty()) {
      task = TaskQueue_.front();
      TaskQueue_.pop_front();
    } else {
      int deadSock = -1;
      for (const auto& [sock, executingTask] : WorkerToTask_) {
        if (!Workers_[sock].Alive) {
          deadSock = sock;
          task = executingTask;
          break;
        }
      }
      if (deadSock == -1) {
        return;
      } else {
        WorkerToTask_.erase(deadSock);
      }
    }

    int workerSock = -1;
    for (const auto& [sock, worker] : Workers_) {
      if (!worker.Alive || WorkerToTask_.contains(sock)) {
        continue;
      }
      workerSock = sock;
      break;
    }
    if (workerSock == -1) {
      TaskQueue_.push_front(task);
      return;
    }

    char code = static_cast<char>(MsgCode::Calc);
    CalcMsg body{task.Interval};
    char calcMsg[sizeof(char) + sizeof(body)];
    calcMsg[0] = code;
    std::memcpy(calcMsg + 1, &body, sizeof(body));

    if (send(workerSock, calcMsg, sizeof(calcMsg), 0) < 0) {
      throw std::runtime_error(std::format("Error while sending CalcMsg"));
    }
    WorkerToTask_[workerSock] = task;
    Log.Debug("Send CalcMsg on interval=({})--({}) to id={}",
              task.Interval.first, task.Interval.second,
              Workers_[workerSock].Key);
  }
}

}  // namespace integral
