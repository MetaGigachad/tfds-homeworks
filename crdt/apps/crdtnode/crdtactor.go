package crdtnode

import (
	opt "chadcrdt/internal/options"
	"fmt"
	"time"

	// "math/rand"
	// "sync"
	// "time"

	. "chadcrdt/internal/utils"

	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/net/edf"
)

func factory_CrdtActor() gen.ProcessBehavior {
	return &CrdtActor{}
}

type CrdtActor struct {
	act.Actor

	clock VectorClock

	cancelRetry map[CancelFuncKey]gen.CancelFunc

	data map[string]CrdtRowValue
}

func (a *CrdtActor) Init(args ...any) error {
	Must(edf.RegisterTypeOf(NodeId(0)))
	Must(edf.RegisterTypeOf(VectorClock{}))
	Must(edf.RegisterTypeOf(CrdtRow{}))
	Must(edf.RegisterTypeOf(NewRowMessage{}))
	Must(edf.RegisterTypeOf(NewRowAckMessage{}))
	a.clock = make(VectorClock, opt.NodeCount)
	a.cancelRetry = make(map[CancelFuncKey]gen.CancelFunc)
	a.data = make(map[string]CrdtRowValue)
	a.Log().Info("started process with name %s", a.Name())
	return nil
}

func (a *CrdtActor) HandleMessage(from gen.PID, message any) error {
	switch msg := message.(type) {
	case NewClientRowMessage:
		return a.HandleNewClientRowMessage(&from, &msg)
	case RetryNewRowMessage:
		return a.HandleRetryNewRowMessage(&from, &msg)
	case NewRowMessage:
		return a.HandleNewRowMessage(&from, &msg)
	case NewRowAckMessage:
		return a.HandleNewRowAckMessage(&from, &msg)
	}

	return nil
}

func (a *CrdtActor) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
	switch req := request.(type) {
	case GetValueRequest:
		return a.HandleGetValue(&from, &req)
	default:
		return false, fmt.Errorf("Invalid request type: %T", request)
	}
}

func (a *CrdtActor) HandleGetValue(from *gen.PID, request *GetValueRequest) (any, error) {
	value, ok := a.data[request.Key]
	if ok {
		return value.Value, nil
	} else {
		return KeyNotFound{}, nil
	}
}

func (a *CrdtActor) HandleNewClientRowMessage(from *gen.PID, message *NewClientRowMessage) error {
	a.clock.Inc()
	msg := NewRowMessage{From: NodeId(opt.NodeId), Row: CrdtRow{Key: message.Key, Value: message.Value, Timestamp: a.clock}}
	a.ApplyNewRow(msg.From, &msg.Row)
	for i := 1; i <= opt.NodeCount; i++ {
		if i == opt.NodeId {
			continue
		}
		a.ReliableSend(NodeId(i), &msg)
	}
	return nil
}

func (a *CrdtActor) HandleRetryNewRowMessage(from *gen.PID, message *RetryNewRowMessage) error {
	err := a.Send(a.SelfOnNode(message.To), message.Msg)
	if err != nil {
		a.Log().Warning("Retry of NewRowMessage to node %d failed: %s", message.To, err)
	}
	a.ScheduleRetry(message.To, &message.Msg)
	return nil
}

func (a *CrdtActor) HandleNewRowMessage(from *gen.PID, message *NewRowMessage) error {
	a.clock.Sync(&message.Row.Timestamp)
	a.ApplyNewRow(message.From, &message.Row)
    err := a.Send(a.SelfOnNode(message.From), NewRowAckMessage{SelfId: message.Row.Timestamp[message.From - 1], From: NodeId(opt.NodeId)})
	if err != nil {
		a.Log().Warning("Sending of NewRowAckMessage to node %d failed: %s", message.From, err)
	}
	return nil
}

func (a *CrdtActor) HandleNewRowAckMessage(from *gen.PID, message *NewRowAckMessage) error {
    key := CancelFuncKey{SelfId: message.SelfId, To: message.From}
    if cancel, ok := a.cancelRetry[key]; ok {
        cancel()
        delete(a.cancelRetry, key)
    }
	return nil
}

func (a *CrdtActor) ApplyNewRow(from NodeId, message *CrdtRow) {
    ts := make(VectorClock, opt.NodeCount)
    copy(ts, message.Timestamp)
    newValue := CrdtRowValue{Value: message.Value, Timestamp: ts}

	val, ok := a.data[message.Key]
	if !ok {
		a.data[message.Key] = newValue
		return
	}

	switch val.Timestamp.Compare(&message.Timestamp) {
	case Before:
		a.data[message.Key] = newValue
	case Equal:
		panic("Impossible situation")
	case Conflict:
		if from < NodeId(opt.NodeId) {
			a.data[message.Key] = newValue
		}
	}
}

func (a *CrdtActor) ReliableSend(to NodeId, msg *NewRowMessage) {
    a.Log().Info("ReliableSend of new pair %s to %s at %v", msg.Row.Key, msg.Row.Value, msg.Row.Timestamp)
	err := a.Send(a.SelfOnNode(to), (*msg))
	if err != nil {
		a.Log().Warning("Sending NewRowMessage to node %d failed: %s", to, err)
	}
	a.ScheduleRetry(to, msg)
}

func (a *CrdtActor) ScheduleRetry(to NodeId, message *NewRowMessage) {
	cancel := Must1(a.SendAfter(a.Name(), RetryNewRowMessage{To: to, Msg: *message}, 1 * time.Second))
    a.cancelRetry[CancelFuncKey{SelfId: message.Row.Timestamp.Self(), To: to}] = cancel
}

func (a *CrdtActor) SelfOnNode(id NodeId) gen.ProcessID {
	return gen.ProcessID{Name: "crdtactor", Node: gen.Atom(opt.MakeNodeName(int(id)))}
}

type GetValueRequest struct {
	Key string
}

type NewClientRowMessage struct {
	Key   string
	Value string
}

type RetryNewRowMessage struct {
	To  NodeId
	Msg NewRowMessage
}

type NewRowMessage struct {
	From NodeId
	Row  CrdtRow
}

type NewRowAckMessage struct {
	SelfId int
    From NodeId
}

type NodeId int

type KeyNotFound struct {
}

type CrdtRow struct {
	Key       string
	Value     string
	Timestamp VectorClock
}

type CrdtRowValue struct {
	Value     string
	Timestamp VectorClock
}

type CancelFuncKey struct {
    SelfId int
    To NodeId
}

type VectorClock []int

type VectorClockCompareResult int

const (
	Before VectorClockCompareResult = iota
	Equal
	After
	Conflict
)

func (c *VectorClock) Compare(o *VectorClock) VectorClockCompareResult {
	beforeCount := 0
	afterCount := 0
	for i := range *c {
		if (*c)[i] <= (*o)[i] {
			beforeCount++
		}
		if (*c)[i] >= (*o)[i] {
			afterCount++
		}
	}
	if beforeCount == len(*c) && afterCount == len(*c) {
		return Equal
	} else if beforeCount == len(*c) {
		return Before
	} else if afterCount == len(*c) {
		return After
	} else {
		return Conflict
	}
}

func (c *VectorClock) Self() int {
	return (*c)[opt.NodeId - 1]
}

func (c *VectorClock) Inc() {
	(*c)[opt.NodeId - 1]++
}

func (c *VectorClock) Sync(o *VectorClock) {
	for i := range *c {
		(*c)[i] = max((*c)[i], (*o)[i])
	}
}
