package crdtnode

import (
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factory_CrdtSup() gen.ProcessBehavior {
	return &CrdtSup{}
}

type CrdtSup struct {
	act.Supervisor
}

// Init invoked on a spawn Supervisor process. This is a mandatory callback for the implementation
func (sup *CrdtSup) Init(args ...any) (act.SupervisorSpec, error) {
	var spec act.SupervisorSpec

	// set supervisor type
	spec.Type = act.SupervisorTypeOneForOne

	// add children
	spec.Children = []act.SupervisorChildSpec{
		{
			Name:    "crdtactor",
			Factory: factory_CrdtActor,
		},
	}

	// set strategy
	spec.Restart.Strategy = act.SupervisorStrategyTransient
	spec.Restart.Intensity = 2 // How big bursts of restarts you want to tolerate.
	spec.Restart.Period = 5    // In seconds.

	return spec, nil
}
