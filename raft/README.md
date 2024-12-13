## Project: "ChadDB"

### Generated with
 - Types for the network messaging: false
 - Enabled Observer (http://localhost:9911): true

### Supervision Tree

Applications
 - `DbNode{}` chaddb/apps/dbnode/dbnode.go
   - `RaftSup{}` chaddb/apps/dbnode/raftsup.go
     - `RaftActor{}` chaddb/apps/dbnode/raftactor.go

Process list that is starting by node directly
 - `HttpApi{}` chaddb/cmd/httpapi.go


#### Used command

This project has been generated with the `ergo` tool. To install this tool, use the following command:

`$ go install ergo.services/tools/ergo@latest`

Below the command that was used to generate this project:

```$ ergo -init ChadDB -with-app DbNode -with-sup DbNode:RaftSup -with-actor RaftSup:RaftActor -with-web HttpApi -with-observer ```
