package node

type RequestVoteArgs struct {
	Term        int
	CandidateId string
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term     int
	LeaderId string
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

type RPC interface {
	SendRequestVote(to string, args RequestVoteArgs) RequestVoteReply
	SendAppendEntries(to string, args AppendEntriesArgs) AppendEntriesReply
}
