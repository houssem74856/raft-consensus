package main

import (
	"fmt"
	"os"
	"raft-consensus/node"
	"strings"
)

/*func main() {
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
}*/

func main() {
	id := os.Getenv("NODE_ID")
	grpcPort := os.Getenv("GRPC_PORT")
	httpPort := os.Getenv("HTTP_PORT")
	peersEnv := os.Getenv("PEERS")

	if id == "" || grpcPort == "" || httpPort == "" || peersEnv == "" {
		fmt.Println("NODE_ID, GRPC_PORT, HTTP_PORT and PEERS env vars are required")
		os.Exit(1)
	}

	peers := map[string]string{}
	for _, p := range strings.Split(peersEnv, ",") {
		parts := strings.SplitN(p, ":", 2)
		peers[parts[0]] = parts[0] + ":" + parts[1]
	}

	peerIds := []string{}
	for pid := range peers {
		peerIds = append(peerIds, pid)
	}

	rpc := node.NewGrpcClients(peers)

	n := node.NewNode(id, peerIds, rpc)

	grpcPortInt := 0
	fmt.Sscanf(grpcPort, "%d", &grpcPortInt)

	httpPortInt := 0
	fmt.Sscanf(httpPort, "%d", &httpPortInt)

	n.StartGrpcServer(grpcPortInt)
	n.StartHttpServer(httpPortInt)
	n.Start()

	select {}
}
