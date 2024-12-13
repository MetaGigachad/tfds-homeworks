package dbnode

import (
	opt "chadcrdt/internal/options"
	"math/rand"
	"sync"
	"time"

	. "chadcrdt/internal/utils"
	"ergo.services/ergo/act"
	"ergo.services/ergo/gen"
	"ergo.services/ergo/net/edf"
)

func factory_RaftActor() gen.ProcessBehavior {
	return &RaftActor{}
}

type RaftActor struct {
	act.Actor
	role           Role
	term           int
	cancelElection *gen.CancelFunc
	votedFor       int
	lastApplied    int

	commitId       int
	log            []LogEntry
	nodeToCommitId map[int]int

	addEntryQueue []AddEntryQueueEntry
}

type AddEntryQueueEntry struct {
	From gen.PID
	Ref  gen.Ref
	Id   int
}

type Role int

const (
	Follower Role = iota
	Candidate
	Leader
)

type LogEntry struct {
	Id    int
	Term  int
	Key   string
    Tombstone bool
	Value string
}

type ActorMessage int

const (
	StartElection ActorMessage = iota
	SendAppendEntries
)

func (a *RaftActor) Init(args ...any) error {
	a.Log().Info("started process with name %s and args %v", a.Name(), args)

    Must(edf.RegisterTypeOf(LogEntry{}))
    Must(edf.RegisterTypeOf(AppendEntries{}))
    Must(edf.RegisterTypeOf(AppendEntriesResult{}))
	Must(edf.RegisterTypeOf(RequestVote{}))

	a.nodeToCommitId = make(map[int]int)

	a.term = 0
    a.commitId = -1
	a.FollowerInit()

	return nil
}

func (a *RaftActor) FollowerInit() {
    a.Log().Info("Role changed: Follower")
	a.role = Follower

	a.ScheduleElection()
}

func (a *RaftActor) ScheduleElection() {
    if a.role != Follower {
        a.Log().Info("Role changed: Follower")
    }
    a.role = Follower
	if a.cancelElection != nil {
		(*a.cancelElection)()
        // a.Log().Info("Election rescheduled")
	}
	leaderTimeout := time.Duration((rand.Intn(1000) + 10000) * 1000 * 1000)
	cancel := Must1(a.SendAfter(a.PID(), StartElection, leaderTimeout))
	a.cancelElection = &cancel
}

func (a *RaftActor) HandleMessage(from gen.PID, message any) error {
	switch message.(type) {
	case ActorMessage:
		if message == StartElection {
			return a.Election()
		} else if message == SendAppendEntries {
			return a.SendAppendEntries()
		}
	}

	return nil
}

func (a *RaftActor) Election() error {
    a.Log().Info("Started Election")
    a.Log().Info("Role changed: Candidate")
	a.role = Candidate
	a.term++
	a.votedFor = opt.NodeId
	votes := 1

	var wg sync.WaitGroup
	wg.Add(opt.NodeCount - 1)
	for i := 1; i <= opt.NodeCount; i++ {
		if i == opt.NodeId {
			continue
		}
		go func() {
			defer wg.Done()
			res, err := a.Call(gen.ProcessID{
				Name: "raftactor",
				Node: gen.Atom(opt.MakeNodeName(i))},
				RequestVote{Term: a.term, CommitId: a.commitId})
			if err != nil {
				a.Log().Warning("Error while sending RequestVote to node %d: %s", i, err)
			} else if res.(bool) {
				votes += 1
			}
		}()
	}
	wg.Wait()

	if votes <= (opt.NodeCount - votes) {
        a.Log().Info("election failed with %d votes", votes)
		a.ScheduleElection()
		return nil
	}
    a.Log().Info("election success with term %d", a.term)
    a.Log().Info("Role changed: Leader")
    a.role = Leader
    a.ScheduleAppendEntries()

	return nil
}

func (a *RaftActor) ScheduleAppendEntries() error {
	leaderTimeout := time.Duration(3000 * 1000 * 1000)
	Must1(a.SendAfter(a.PID(), SendAppendEntries, leaderTimeout))
	return nil
}

type RequestVote struct {
	NodeId   int
	Term     int
	CommitId int
}

type AppendEntries struct {
	Entries  []LogEntry
	CommitId int
}

type AppendEntriesResult struct {
	CommitId int
}

type AddEntry struct {
	Key   string
	Value string
    Tombstone bool
}

func (a *RaftActor) HandleCall(from gen.PID, ref gen.Ref, request any) (any, error) {
    // a.Log().Info("Handling call")
	switch val := request.(type) {
	case RequestVote:
		return a.RequestVote(from, ref, val)
	case AppendEntries:
        // a.Log().Info("Before handling append entries")
		return a.AppendEntries(from, ref, val)
	case AddEntry:
		return a.AddEntry(from, ref, val)
	}

	return false, nil
}

func (a *RaftActor) RequestVote(from gen.PID, ref gen.Ref, request RequestVote) (bool, error) {
	if a.term < request.Term && a.commitId <= request.CommitId {
		a.term = request.Term
		a.votedFor = request.NodeId
		a.ScheduleElection()
        a.Log().Info("Voted for node %d", request.NodeId)
		return true, nil
	}

    a.Log().Info("Not Voted for node %d", request.NodeId)
	return false, nil
}

func (a *RaftActor) SendAppendEntries() error {
    // a.Log().Info("Sending AppendEntries")
    defer a.ScheduleAppendEntries()

	replicatedTo := 1

	for i := 1; i <= opt.NodeCount; i++ {
		if i == opt.NodeId {
			continue
		}
        // TODO: Parallalize this (just spawining goroutines doesn't work sadly, requires rewrite from rpc calls to async messages)
        res, err := a.Call(gen.ProcessID{
            Name: "raftactor",
            Node: gen.Atom(opt.MakeNodeName(i))},
            AppendEntries{Entries: a.getLogUpdatesForNode(i), CommitId: a.commitId})
        if err != nil {
            a.Log().Warning("Error while sending AppendEntries to node %d: %s", i, err)
        } else if val, ok := res.(AppendEntriesResult); ok {
            a.nodeToCommitId[i] = val.CommitId
            replicatedTo += 1
        }
	}

	if replicatedTo <= (opt.NodeCount - replicatedTo) {
		return nil
	}
	a.MoveStateMachine(len(a.log) - 1)
	return nil
}

func (a *RaftActor) getLogUpdatesForNode(nodeId int) []LogEntry {
	if commitId, ok := a.nodeToCommitId[nodeId]; !ok {
		return []LogEntry{}
	} else {
		return a.log[commitId+1 : min(len(a.log), commitId+1+10)]
	}
}

func (a *RaftActor) AppendEntries(from gen.PID, ref gen.Ref, request AppendEntries) (any, error) {
	a.ScheduleElection()

	for _, newEntry := range request.Entries {
		if newEntry.Id < len(a.log) {
			if a.log[newEntry.Id] == newEntry {
				continue
			} else if a.log[newEntry.Id].Term >= newEntry.Term {
				panic("AppendEntries encountered impossible situation 1")
			} else {
				a.log[newEntry.Id] = newEntry
			}
		} else if newEntry.Id == len(a.log) {
			a.log = append(a.log, newEntry)
		} else {
			panic("AppendEntries encountered impossible situation 2")
		}
	}

	a.MoveStateMachine(min(request.CommitId, len(a.log)-1))

    // a.Log().Info("Handled AppendEntries")
	return AppendEntriesResult{CommitId: a.commitId}, nil
}

func (a *RaftActor) MoveStateMachine(toId int) {
	if a.commitId > toId {
		panic("MoveStateMachine encountered impossible situation 1")
	}

	for _, entry := range a.log[a.commitId + 1 : toId+1] {
		var op any
		if !entry.Tombstone {
			op = StorageSet{Key: entry.Key, Value: entry.Value}
		} else {
			op = StorageDel{Key: entry.Key}
		}
		Must1(a.Call(gen.Atom("storageactor"), op))
        a.Log().Info("Moved state machine to id %d", entry.Id)

        if len(a.addEntryQueue) == 0 {
            continue
        }
        req := a.addEntryQueue[0]
        if req.Id != entry.Id {
            continue  
        }
        a.SendResponse(req.From, req.Ref, true)
	}

	a.commitId = toId
}

func (a *RaftActor) AddEntry(from gen.PID, ref gen.Ref, request AddEntry) (any, error) {
	a.log = append(a.log, LogEntry{Id: len(a.log), Term: a.term, Key: request.Key, Value: request.Value})
	a.addEntryQueue = append(a.addEntryQueue, AddEntryQueueEntry{From: from, Ref: ref, Id: len(a.log) - 1})
	return nil, nil
}
