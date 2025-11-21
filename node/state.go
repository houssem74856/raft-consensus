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

type State struct {
	mu          sync.Mutex
	Role        NodeState
	CurrentTerm int
	VotedFor    string
}

func NewState() *State {
	return &State{
		Role:        Follower,
		CurrentTerm: 0,
		VotedFor:    "",
	}
}
