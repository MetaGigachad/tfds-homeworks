#pragma once

#include <map>
#include <optional>

#include "config.hpp"
#include "integral.hpp"

namespace integral {

class Worker {
 public:
  Worker(const Config& config);

  void Start();

 private:
  void SetupUdp();
  void SetupTcp();
  void SetSocketNonBlocking(int sockfd);
  void RunEventLoop();
  void HandleUdp();
  void HandleTcpConnection();
  void HandleTcpMessage(int clientSock);
  void SendResult();

  const Config& Config_;
  int EpollFd_;
  int UdpFd_;
  int TcpFd_;
  std::map<int, int> ClientSockets_;
  std::optional<ComputeIntervalTask> Task_;
  int MasterSock_;
  time_t StartedTaskAt_; 
};

}  // namespace integral
