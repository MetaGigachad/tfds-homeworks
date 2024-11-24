cmake_minimum_required(VERSION 3.16)
project(integral VERSION 1.0 LANGUAGES CXX)

add_executable(${PROJECT_NAME}-master
    src/args.cpp
    src/args.hpp
    src/config.cpp
    src/config.hpp
    src/integral.cpp
    src/integral.hpp
    src/logger.cpp
    src/logger.hpp
    src/main_master.cpp
    src/server.cpp
    src/server.hpp
)
target_compile_features(${PROJECT_NAME}-master PRIVATE cxx_std_23)

add_executable(${PROJECT_NAME}-worker
    src/args.cpp
    src/args.hpp
    src/config.cpp
    src/config.hpp
    src/integral.cpp
    src/integral.hpp
    src/logger.cpp
    src/logger.hpp
    src/main_worker.cpp
    src/worker.cpp
    src/worker.hpp
)
target_compile_features(${PROJECT_NAME}-worker PRIVATE cxx_std_23)
