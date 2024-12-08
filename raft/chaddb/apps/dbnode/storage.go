package dbnode

import (
	"fmt"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
)

func factory_StorageSup() gen.ProcessBehavior {
	return &StorageSup{}
}

type StorageSup struct {
	act.Supervisor
}

func (sup *StorageSup) Init(args ...any) (act.SupervisorSpec, error) {
	var spec act.SupervisorSpec
	spec.Type = act.SupervisorTypeOneForOne
	spec.Children = []act.SupervisorChildSpec{
		{
			Name:    "storageactor",
			Factory: factory_StorageActor,
		},
	}
	spec.Restart.Strategy = act.SupervisorStrategyTransient
	spec.Restart.Intensity = 2
	spec.Restart.Period = 5
	return spec, nil
}


func factory_StorageActor() gen.ProcessBehavior {
	return &StorageActor{}
}

type StorageActor struct {
	act.Actor
    data map[string]string
}

func (a *StorageActor) Init(args ...any) error {
	a.Log().Info("started process with name %s and args %v", a.Name(), args)
	return nil
}

type StorageGet struct {
    Key string
}

type StorageSet struct {
    Key string
    Value string
}

type StorageDel struct {
    Key string
}

func (a *StorageActor) HandleMessage(from gen.PID, message any) error {
	a.Log().Info("got message from %s", from)
	return nil
}

func (a *StorageActor) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	a.Log().Info("got request from %s with reference %s", from, ref)

    switch request.(type) {
    case StorageGet:
        return a.HandleGet(from, request.(StorageGet));
    case StorageSet:
        return a.HandleSet(from, request.(StorageSet));
    case StorageDel:
        return a.HandleDel(from, request.(StorageDel));
    default:
        return nil, fmt.Errorf("Invalid request type: %T", request);
    }
}

func (a *StorageActor) HandleGet(from gen.PID, message StorageGet) (any, error) {
    value, ok := a.data[message.Key]
    if ok {
        return value, nil
    } else {
        return nil, nil
    }
}

func (a *StorageActor) HandleSet(from gen.PID, message StorageSet) (any, error) {
    a.data[message.Key] = message.Value
    return nil, nil
}

func (a *StorageActor) HandleDel(from gen.PID, message StorageDel) (any, error) {
    delete(a.data, message.Key)
    return nil, nil
}
