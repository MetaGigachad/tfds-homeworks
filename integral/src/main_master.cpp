#include "config.hpp"
#include "args.hpp"
#include "server.hpp"

int main(int argc, char* argv[]) {
  auto args = integral::ReadArgs(argc, argv);
  if (!args.has_value()) {
    return 1;
  }
  auto config = integral::ReadConfig(args->StaticConfigPath);

  auto server = integral::Server(config);
  server.Start();

  return 0;
}
