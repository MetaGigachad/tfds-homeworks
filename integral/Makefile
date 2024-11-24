CMAKE_FLAGS ?= -DCMAKE_EXPORT_COMPILE_COMMANDS=ON

.PHONY: cmake-debug cmake-release
cmake-debug cmake-release: cmake-%:
	cmake -S . -B build_$* -DCMAKE_BUILD_TYPE=$(shell echo $* | sed 's/.*/\u&/') $(CMAKE_FLAGS)

.PHONY: build-debug build-release
build-debug build-release: build-%: cmake-%
	cmake --build build_$*

.PHONY: run-master-debug run-master-release
run-master-debug run-master-release: run-master-%:
	./build_$*/integral-master --static-config ./configs/master.env

.PHONY: run-worker-debug run-worker-release
run-worker-debug run-worker-release: run-worker-%:
	./build_$*/integral-worker --static-config ./configs/worker.env

.PHONY: docker-build
docker-build:
	docker build -t integral-master:1.0.0 . -f ./docker/master.dockerfile --shm-size=4g
	docker build -t integral-worker:1.0.0 . -f ./docker/worker.dockerfile --shm-size=4g
