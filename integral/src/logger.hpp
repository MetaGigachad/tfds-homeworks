#pragma once

#include <chrono>
#include <format>
#include <iostream>
#include <mutex>
#include <print>
#include <string>

class Logger {
 public:
  enum class Level { DEBUG, INFO, WARNING, ERROR };

  Logger(Level min_level = Level::INFO) : MinLevel_(min_level) {}

  void Log(Level level, std::string_view fmt, std::format_args&& args) {
    if (level < MinLevel_) return;

    std::lock_guard<std::mutex> lock(Mutex_);
    std::string message =
        std::vformat(fmt, args);
    std::string log_entry =
        std::format(R"(ts="{}" level="{}" msg="{}")", CurrentTime(),
                    LevelToString(level), message);
    std::cout << log_entry << std::endl;
  }

  template <typename... Args>
  void Info(std::string_view fmt, Args&&... args) {
    Log(Level::INFO, fmt, std::make_format_args(args...));
  }

  template <typename... Args>
  void Warning(std::string_view fmt, Args&&... args) {
    Log(Level::WARNING, fmt, std::make_format_args(args...));
  }

  template <typename... Args>
  void Error(std::string_view fmt, Args&&... args) {
    Log(Level::ERROR, fmt, std::make_format_args(args...));
  }

  template <typename... Args>
  void Debug(std::string_view fmt, Args&&... args) {
    Log(Level::DEBUG, fmt, std::make_format_args(args...));
  }

 private:
  Level MinLevel_;
  std::mutex Mutex_;

  static std::string CurrentTime() {
    auto now = std::chrono::system_clock::now();
    return std::format("{:%Y-%m-%d %H:%M:%S}", now);
  }

  static std::string LevelToString(Level level) {
    switch (level) {
      case Level::INFO:
        return "INFO";
      case Level::WARNING:
        return "WARNING";
      case Level::ERROR:
        return "ERROR";
      case Level::DEBUG:
        return "DEBUG";
      default:
        return "UNKNOWN";
    }
  }
};

extern Logger Log;
