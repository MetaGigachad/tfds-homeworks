# ChadDB: In-memory key-value storage with raft consensus

### How to run

Run example for 3 nodes with:

```bash
go run ./cmd -node-id 1 -node-count 3 -cookie somecommonhash &
go run ./cmd -node-id 2 -node-count 3 -cookie somecommonhash &
go run ./cmd -node-id 3 -node-count 3 -cookie somecommonhash
```

Use `-help` to learn about other arguments.

### Utilities

Interact with replicas using `./chadcli`:

```
Usage: ./chadcli <port> {get|set|del} <key> [value]
```
