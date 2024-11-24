#pragma once

#include <functional>

namespace integral {

using Fn = std::function<double(double)>;

struct ComputeIntervalTask {
  std::pair<double, double> Interval;
};

double ComputeInterval(Fn f, const ComputeIntervalTask& task,
                       unsigned int subdivisionCount);

std::vector<ComputeIntervalTask> Split(const ComputeIntervalTask& task, unsigned int n);

}  // namespace intergral
