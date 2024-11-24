#include "worker.hpp"

#include <fcntl.h>
#include <netinet/in.h>
#include <sys/epoll.h>
#include <sys/socket.h>
#include <unistd.h>

#include <cstring>
#include <stdexcept>

#include "logger.hpp"
#include "msg.hpp"

namespace integral {

Worker::Worker(const Config& config)
    : Config_(config), EpollFd_(-1), UdpFd_(-1), TcpFd_(-1) {
  EpollFd_ = epoll_create1(0);
  if (EpollFd_ == -1) {
    throw std::runtime_error("Failed to create epoll file descriptor");
  }
}

void Worker::Start() {
  Log.Info("Starting worker");

  SetupUdp();
  SetupTcp();
  RunEventLoop();
}

void Worker::SetupUdp() {
  UdpFd_ = socket(AF_INET, SOCK_DGRAM, 0);
  if (UdpFd_ == -1) {
    throw std::runtime_error("Failed to create UDP socket");
  }

  sockaddr_in udpAddr{};
  udpAddr.sin_family = AF_INET;
  udpAddr.sin_port = htons(std::stoul(Config_.at("udpPort")));
  udpAddr.sin_addr.s_addr = INADDR_ANY;

  if (bind(UdpFd_, (struct sockaddr*)&udpAddr, sizeof(udpAddr)) == -1) {
    throw std::runtime_error("Failed to bind UDP socket");
  }

  SetSocketNonBlocking(UdpFd_);

  epoll_event udpEvent{};
  udpEvent.events = EPOLLIN;
  udpEvent.data.fd = UdpFd_;
  if (epoll_ctl(EpollFd_, EPOLL_CTL_ADD, UdpFd_, &udpEvent) == -1) {
    throw std::runtime_error("Failed to add UDP socket to epoll");
  }
}

void Worker::SetupTcp() {
  TcpFd_ = socket(AF_INET, SOCK_STREAM, 0);
  if (TcpFd_ == -1) {
    throw std::runtime_error("Failed to create TCP socket");
  }

  sockaddr_in tcpAddr{};
  tcpAddr.sin_family = AF_INET;
  tcpAddr.sin_port = htons(std::stoul(Config_.at("tcpPort")));
  tcpAddr.sin_addr.s_addr = INADDR_ANY;

  if (bind(TcpFd_, (struct sockaddr*)&tcpAddr, sizeof(tcpAddr)) == -1) {
    throw std::runtime_error("Failed to bind TCP socket");
  }

  if (listen(TcpFd_, 10) == -1) {
    throw std::runtime_error("Failed to listen on TCP socket");
  }

  SetSocketNonBlocking(TcpFd_);

  epoll_event tcpEvent{};
  tcpEvent.events = EPOLLIN | EPOLLET;
  tcpEvent.data.fd = TcpFd_;
  if (epoll_ctl(EpollFd_, EPOLL_CTL_ADD, TcpFd_, &tcpEvent) == -1) {
    throw std::runtime_error("Failed to add TCP socket to epoll");
  }
}

void Worker::SetSocketNonBlocking(int sockfd) {
  int flags = fcntl(sockfd, F_GETFL, 0);
  if (flags == -1) {
    throw std::runtime_error("Failed to get socket flags");
  }

  flags |= O_NONBLOCK;
  if (fcntl(sockfd, F_SETFL, flags) == -1) {
    throw std::runtime_error("Failed to set socket non-blocking");
  }
}

void Worker::RunEventLoop() {
  const int maxEvents = 10;
  epoll_event events[maxEvents];

  while (true) {
    try {
      int nEvents = epoll_wait(EpollFd_, events, maxEvents, 50);
      if (nEvents == -1) {
        throw std::runtime_error("epoll_wait failed");
      }

      for (int i = 0; i < nEvents; ++i) {
        if (events[i].data.fd == UdpFd_) {
          HandleUdp();
        } else if (events[i].data.fd == TcpFd_) {
          HandleTcpConnection();
        } else {
          HandleTcpMessage(events[i].data.fd);
        }
      }

      if (difftime(time(NULL), StartedTaskAt_) >= 10) {
        SendResult();
      }
    } catch (const std::runtime_error& err) {
      Log.Error("Error occured in epoll event loop: {}", err.what());
    }
  }
}

void Worker::HandleUdp() {
  char buffer[1024];
  sockaddr_in clientAddr{};
  socklen_t clientAddrLen = sizeof(clientAddr);

  int len = recvfrom(UdpFd_, buffer, sizeof(buffer), 0,
                     (struct sockaddr*)&clientAddr, &clientAddrLen);
  if (len == -1) {
    throw std::runtime_error("Failed to receive UDP message");
  }

  char msg[] = {static_cast<char>(MsgCode::HelloAck)};
  if (buffer[0] == static_cast<char>(MsgCode::Hello)) {
    sendto(UdpFd_, msg, sizeof(msg), 0, (struct sockaddr*)&clientAddr,
           clientAddrLen);
    Log.Debug("Sent HelloAck");
  }
}

void Worker::HandleTcpConnection() {
  sockaddr_in clientAddr{};
  socklen_t clientAddrLen = sizeof(clientAddr);
  int clientSock =
      accept(TcpFd_, (struct sockaddr*)&clientAddr, &clientAddrLen);

  if (clientSock == -1) {
    throw std::runtime_error("Failed to accept TCP connection");
  }

  SetSocketNonBlocking(clientSock);

  epoll_event event{};
  event.events = EPOLLIN | EPOLLET;
  event.data.fd = clientSock;
  if (epoll_ctl(EpollFd_, EPOLL_CTL_ADD, clientSock, &event) == -1) {
    throw std::runtime_error("Failed to add client socket to epoll");
  }

  ClientSockets_[clientSock] = clientSock;
  Log.Debug("Connected to master");
}

void Worker::HandleTcpMessage(int clientSock) {
  char buffer[1024];
  int len = read(clientSock, buffer, sizeof(buffer));
  if (len == -1) {
    throw std::runtime_error("Failed to read TCP message");
  }

  char msg[] = {static_cast<char>(MsgCode::PingAck)};
  if (buffer[0] == static_cast<char>(MsgCode::Ping)) {
    Log.Debug("Respoding to ping");
    write(clientSock, msg, sizeof(msg));
  } else if (buffer[0] == static_cast<char>(MsgCode::Calc)) {
    CalcMsg body;
    std::memcpy(&body, buffer + 1, sizeof(body));
    if (Task_) {
      throw std::runtime_error(
          "Received new Task, but still calculating old one");
    }
    Task_ = ComputeIntervalTask{body.Range};
    StartedTaskAt_ = time(NULL);
    MasterSock_ = clientSock;
    Log.Debug("Got CalcMsg");
  } else {
    throw std::runtime_error("Received unknown message type over TCP");
  }
}

void Worker::SendResult() {
  if (!Task_) {
    return;
  }

  CalcResult result{
      Task_->Interval,
      ComputeInterval([](double x) { return x + 1; }, Task_.value(), 10)};
  Task_ = std::nullopt;

  char code = static_cast<char>(MsgCode::CalcResult);
  char calcResult[sizeof(code) + sizeof(result)];
  calcResult[0] = code;
  std::memcpy(calcResult + 1, &result, sizeof(result));
  if (send(MasterSock_, calcResult, sizeof(calcResult), 0) < 0) {
    throw std::runtime_error(std::format("Error while sending CalcResult"));
  }
  Log.Debug("Send CalcResult");
}

}  // namespace integral
