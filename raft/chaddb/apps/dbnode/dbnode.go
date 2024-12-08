package dbnode

import (
	"ergo.services/ergo/gen"
)

func CreateDbNode() gen.ApplicationBehavior {
	return &DbNode{}
}

type DbNode struct{}

// Load invoked on loading application using method ApplicationLoad of gen.Node interface.
func (app *DbNode) Load(node gen.Node, args ...any) (gen.ApplicationSpec, error) {
	return gen.ApplicationSpec{
		Name:        "dbnode",
		Description: "description of this application",
		Mode:        gen.ApplicationModeTransient,
		Group: []gen.ApplicationMemberSpec{
			{
				Name:    "raftsup",
				Factory: factory_RaftSup,
			},
			{
				Name:    "storagesup",
				Factory: factory_StorageSup,
			},
		},
	}, nil
}

// Start invoked once the application started
func (app *DbNode) Start(mode gen.ApplicationMode) {}

// Terminate invoked once the application stopped
func (app *DbNode) Terminate(reason error) {}
