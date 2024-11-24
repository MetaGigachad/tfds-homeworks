#pragma once

#include <optional>
#include <string>

namespace integral {

struct Args {
  std::string StaticConfigPath;
};

std::optional<Args> ReadArgs(int argc, char* argv[]);

}  // namespace integral
