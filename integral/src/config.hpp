#pragma once

#include <string>
#include <unordered_map>

namespace integral {

using Config = std::unordered_map<std::string, std::string>;

Config ReadConfig(
    const std::string &filename);

}
