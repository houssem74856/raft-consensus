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
	mu   sync.Mutex
	Role NodeState
}

func NewState() *State {
	return &State{
		Role: Follower,
	}
}

func (s *State) BecomeFollower(term int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Role = Follower
}

func (s *State) BecomeCandidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Role = Candidate
}

func (s *State) BecomeLeader() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Role = Leader
}
