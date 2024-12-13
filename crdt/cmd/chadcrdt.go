package main

import (
	"flag"

	"chadcrdt/apps/dbnode"
	opt "chadcrdt/internal/options"

	"ergo.services/application/observer"

	"ergo.services/ergo"
	"ergo.services/ergo/gen"
)

func main() {
	var options gen.NodeOptions

	flag.Parse()
	opt.ObserverPort = 4000 + opt.NodeId
	opt.ApiPort = 5000 + opt.NodeId

	// create applications that must be started
	apps := []gen.ApplicationBehavior{
		observer.CreateApp(observer.Options{Port: uint16(opt.ObserverPort)}),
		dbnode.CreateDbNode(),
	}
	options.Applications = apps

	// set network options
	options.Network.Cookie = opt.NodeCookie

	// starting node
	node, err := ergo.StartNode(gen.Atom(opt.MakeNodeName(opt.NodeId)), options)
	if err != nil {
		panic(err)
	}

	// starting process HttpApi
	if _, err := node.SpawnRegister("httpapi", factory_HttpApi, gen.ProcessOptions{}); err != nil {
		panic(err)
	}

	node.Wait()
}
