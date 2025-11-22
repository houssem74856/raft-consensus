package main

import (
	"simple-go-consensus/node"
)

func main() {
	port := 3000
	ids := []string{"n1", "n2", "n3", "n4"}

	rpc := node.NewInMemoryRPC()

	var nodes []*node.Node

	for _, id := range ids {
		peers := []string{}
		for _, other := range ids {
			if other != id {
				peers = append(peers, other)
			}
		}

		n := node.NewNode(id, peers, rpc)
		rpc.Register(n)
		nodes = append(nodes, n)
	}

	for _, n := range nodes {
		n.Start()
		n.StartServer(port)
		port++
	}

	select {} // blocks forever to keep program running while nodes operate
}
