package options

import (
	"flag"
	"fmt"

	"ergo.services/ergo/lib"
)

var (
	NodeId       int
	NodeCookie   string
	ObserverPort int
	ApiPort      int
	NodeCount    int
)

func init() {
	flag.IntVar(&NodeId, "node-id", 1, "node id")
	flag.StringVar(&NodeCookie, "cookie", lib.RandomString(16), "a secret cookie for the network messaging")
	flag.IntVar(&ObserverPort, "observer-port", 4000, "port for observer")
	flag.IntVar(&ApiPort, "api-port", 5000, "port for api")
	flag.IntVar(&NodeCount, "node-count", 3, "amount of replicas")
}

func MakeNodeName(id int) string {
    return fmt.Sprintf("chaddb-node-%d@localhost", id)
}
