FROM gcc:14.2

RUN apt-get -y update && apt-get -y install cmake

WORKDIR /app
COPY . /app

RUN make build-release

ENTRYPOINT ["./build_release/integral-master", "--static-config", "./configs/master.env"]
