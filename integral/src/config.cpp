#include <fstream>
#include <string>
#include <unordered_map>

#include "logger.hpp"

namespace integral {

std::unordered_map<std::string, std::string> ReadConfig(
    const std::string &filename) {
  std::unordered_map<std::string, std::string> config;
  std::ifstream file(filename);
  if (!file) {
    throw std::runtime_error("Unable to open static config file: " + filename);
  }
  Log.Info("Loading config");

  std::string line;
  while (std::getline(file, line)) {
    if (line.empty() || line[0] == '#') continue;

    auto delimiterPos = line.find('=');
    if (delimiterPos != std::string::npos) {
      std::string key = line.substr(0, delimiterPos);
      std::string value = line.substr(delimiterPos + 1);
      config[key] = value;
      Log.Debug("Loaded config option: {}={}", key, value);
    }
  }
  return config;
}

};  // namespace integral
