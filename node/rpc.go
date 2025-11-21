package node

type RequestVoteArgs struct {
	Term        int
	CandidateId string
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type RPC interface {
	SendRequestVote(to string, args RequestVoteArgs) RequestVoteReply
}
