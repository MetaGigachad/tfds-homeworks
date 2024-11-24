# Integral

This solution implemets two services: master and worker. Master has to calculate integral
of a function by separating interval into segments and giving them to worker services.
Any case of failure of worker is handled.

### How to build

To build docker images run this

```sh
make docker-build
```

### How to run

Run example using docker compose

```sh
docker compose -f ./docker/compose.yaml up
```
