package node

import (
	"sync"
)

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

type LogEntry struct {
	Term    int
	Command Command
}

type State struct {
	mu          sync.Mutex
	Role        NodeState
	CurrentTerm int
	VotedFor    string
	Log         []LogEntry
	CommitIndex int
	LastApplied int
}

func NewState() *State {
	return &State{
		Role:        Follower,
		CurrentTerm: 0,
		VotedFor:    "",
		CommitIndex: -1,
		LastApplied: -1,
	}
}
