package node

import (
	"sync"
)

type StateMachine struct {
	mu    sync.Mutex
	store map[string]string
}

func NewStateMachine() *StateMachine {
	return &StateMachine{
		store: make(map[string]string),
	}
}

func (sm *StateMachine) Apply(entry LogEntry) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.store[entry.Command.Key] = entry.Command.Value
}

func (sm *StateMachine) Get(key string) (string, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	val, ok := sm.store[key]
	return val, ok
}
