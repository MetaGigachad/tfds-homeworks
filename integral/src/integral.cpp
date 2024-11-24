#include "integral.hpp"

namespace integral {

double ComputeInterval(Fn f, const ComputeIntervalTask& task,
                       unsigned int subdivisionCount) {
  double step = (task.Interval.second - task.Interval.first) / subdivisionCount;
  double integral = 0.0;

  for (int i = 0; i < subdivisionCount; ++i) {
    double x1 = task.Interval.first + i * step;
    double x2 = x1 + step;
    integral += 0.5 * (f(x1) + f(x2)) * step;  // Trapezoidal Rule
  }

  return integral;
}

std::vector<ComputeIntervalTask> Split(const ComputeIntervalTask& task, unsigned int n) {
  std::vector<ComputeIntervalTask> tasks;
  double step = (task.Interval.second - task.Interval.first) / n;

  for (int i = 0; i < n; ++i) {
    double x1 = task.Interval.first + i * step;
    double x2 = x1 + step;
    tasks.push_back(ComputeIntervalTask{{x1, x2}});
  }

  return tasks;
}

}  // namespace intergral
