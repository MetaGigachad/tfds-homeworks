# ChadCRDT: conflict-free replicated data type for key value storage

### How to run

Run example for 3 nodes with:

```bash
go run ./cmd -node-id 1 -node-count 3 -cookie somecommonhash &
go run ./cmd -node-id 2 -node-count 3 -cookie somecommonhash &
go run ./cmd -node-id 3 -node-count 3 -cookie somecommonhash
```

Use `-help` to learn about other arguments.

### Utilities

Interact with replicas using `./crdtcli`:

```
Usage: ./crdtcli <port> {get|set|op} <key> [value]
```

### Tests

Run tests with `./test`:
```
Usage: ./test <test-name>
```
