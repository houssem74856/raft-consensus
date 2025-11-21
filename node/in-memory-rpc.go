package node

type InMemoryRPC struct {
	nodes map[string]*Node
}

func NewInMemoryRPC() *InMemoryRPC {
	return &InMemoryRPC{
		nodes: make(map[string]*Node),
	}
}

func (r *InMemoryRPC) Register(n *Node) {
	r.nodes[n.Id] = n
}

func (r *InMemoryRPC) SendRequestVote(to string, args RequestVoteArgs) RequestVoteReply {
	target := r.nodes[to]
	return target.HandleRequestVote(args)
}
