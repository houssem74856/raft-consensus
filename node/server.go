package node

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Command struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (n *Node) StartServer(port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("/command", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var cmd Command
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid JSON"))
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

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	fmt.Println("Node", n.Id, "listening on", addr)

	go func() {
		if err := http.ListenAndServe(addr, mux); err != nil {
			fmt.Println("Server error on port", port, ":", err)
		}
	}()
}
