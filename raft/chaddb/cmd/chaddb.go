package main

import (
	"flag"

	"chaddb/apps/dbnode"

	"ergo.services/application/observer"

	"ergo.services/ergo"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/lib"
)

var (
	OptionNodeName   string
	OptionNodeCookie string
)

func init() {
	flag.StringVar(&OptionNodeName, "name", "ChadDB@localhost", "node name")
	flag.StringVar(&OptionNodeCookie, "cookie", lib.RandomString(16), "a secret cookie for the network messaging")
}

func main() {
	var options gen.NodeOptions

	flag.Parse()

	// create applications that must be started
	apps := []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{}),
		dbnode.CreateDbNode(),
	}
	options.Applications = apps

	// set network options
	options.Network.Cookie = OptionNodeCookie

	// starting node
	node, err := ergo.StartNode(gen.Atom(OptionNodeName), options)
	if err != nil {
		panic(err)
	}

	// starting process HttpApi
	if _, err := node.SpawnRegister("httpapi", factory_HttpApi, gen.ProcessOptions{}); err != nil {
		panic(err)
	}

	node.Wait()
}
