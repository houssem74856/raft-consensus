package node

import (
	"fmt"
	"time"
)

type Node struct {
	Id    string
	State *State
}

func NewNode(id string) *Node {
	return &Node{
		Id:    id,
		State: NewState(),
	}
}

func (n *Node) Start() {
	fmt.Println("node ", n.Id, " started")
	for {
		time.Sleep(100 * time.Millisecond)
	}
}
