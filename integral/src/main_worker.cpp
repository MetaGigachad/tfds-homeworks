#include "config.hpp"
#include "args.hpp"
#include "worker.hpp"

int main(int argc, char* argv[]) {
  auto args = integral::ReadArgs(argc, argv);
  if (!args.has_value()) {
    return 1;
  }
  auto config = integral::ReadConfig(args->StaticConfigPath);

  auto worker = integral::Worker(config);
  worker.Start();

  return 0;
}
