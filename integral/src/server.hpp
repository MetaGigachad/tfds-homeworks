#pragma once

#include <netinet/in.h>
#include <deque>
#include <set>
#include <unordered_set>

#include "config.hpp"
#include "integral.hpp"

namespace integral {

struct WorkerConnection {
  bool Alive;
  bool RecievedLastPingAck;
  std::string Key;
};

using WorkerMap = std::unordered_map<int, WorkerConnection>;
using TaskQueue = std::deque<ComputeIntervalTask>;
using TaskMap = std::unordered_map<int, ComputeIntervalTask>;

class Server {
 public:
  explicit Server(const Config& config);

  void Start();

 private:
  int MakeUdpSocket();
  sockaddr_in MakeBroadcastAddr();
  void BroadcastHello(int sock, sockaddr_in broadcastAddr);
  void EnqueueCalculation();

  void EpollEventLoop(int udpSock);

  void HandleUdpMessage(int sock);
  void ConnectToWorker(sockaddr_in& recvAddr);

  void HandleTcpMessage(int tcpSock);

  void PingWorkers();
  void AssignCalc();

  const Config& Config_;
  WorkerMap Workers_;
  std::unordered_set<std::string> WorkerAddrs_;
  TaskQueue TaskQueue_;
  TaskMap WorkerToTask_;


  int EpollFd_;

  std::set<std::pair<double, double>> CalculatedIntervals_;
  double Result_;
};

}  // namespace integral
