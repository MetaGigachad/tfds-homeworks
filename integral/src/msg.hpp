#pragma once

#include <utility>

namespace integral {

enum class MsgCode {
  Hello = 'h',
  HelloAck = 'H',
  Ping = 'p',
  PingAck = 'P',
  Calc = 'c',
  CalcResult = 'C',
};

struct CalcMsg {
  std::pair<double, double> Range;
};

struct CalcResult {
  std::pair<double, double> Range;
  double Result;
};

}  // namespace integral
