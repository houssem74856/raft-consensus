package node

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	pb "raft-consensus/proto"

	"google.golang.org/grpc"
)

type Command struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (n *Node) StartGrpcServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		fmt.Println("grpc listening error on port", port, ":", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRaftServer(grpcServer, NewRaftServer(n))

	fmt.Println("node", n.Id, "grpc listening on port", port)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Println("grpc server error on port", port, ":", err)
		}
	}()
}

func (n *Node) StartHttpServer(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/command", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var cmd Command
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid json"))
			return
		}

		n.State.mu.Lock()
		role := n.State.Role
		n.State.mu.Unlock()
		if role != Leader {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte("not leader, current leader unknown"))
			return
		}

		entry := LogEntry{
			Term:    n.State.CurrentTerm,
			Command: cmd,
		}

		n.State.mu.Lock()
		n.State.Log = append(n.State.Log, entry)
		n.State.mu.Unlock()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("entry appended"))
	})

	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val, ok := n.SM.Get(key)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("key not found"))
			return
		}
		w.Write([]byte(val))
	})

	fmt.Println("node", n.Id, "http listening on", fmt.Sprintf(":%d", port))

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
			fmt.Println("http server error on port", port, ":", err)
		}
	}()
}
