#include "args.hpp"

#include <getopt.h>

#include <iostream>
#include <optional>
#include <string>

namespace integral {

namespace {

void PrintUsage(const char* programName) {
  std::cerr << "Usage: " << programName << " --static-config PATH\n";
}

bool ReadArgs(int argc, char* argv[], Args& args) {
  const char* const shortOpts = "";
  const option longOpts[] = {{"static-config", required_argument, nullptr, 's'},
                             {nullptr, 0, nullptr, 0}};

  while (true) {
    const int opt = getopt_long(argc, argv, shortOpts, longOpts, nullptr);
    if (opt == -1) {
      break;
    }

    switch (opt) {
      case 's':
        args.StaticConfigPath = optarg;
        break;
      case '?':
      default:
        return false;
    }
  }

  return true;
}

bool ValidateArgs(const Args& args) {
  if (args.StaticConfigPath.empty()) {
    std::cerr << "Error: --static-config PATH is required\n";
    return false;
  }
  return true;
}

}  // namespace

std::optional<Args> ReadArgs(int argc, char* argv[]) {
  Args args;

  if (!ReadArgs(argc, argv, args)) {
    std::cerr << "Error: Invalid arguments\n";
    PrintUsage(argv[0]);
    return std::nullopt;
  }

  if (!ValidateArgs(args)) {
    PrintUsage(argv[0]);
    return std::nullopt;
  }

  return args;
}

}  // namespace integral
