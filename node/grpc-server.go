package node

import (
	"context"
	pb "raft-consensus/proto"
)

type RaftServer struct {
	pb.UnimplementedRaftServer
	node *Node
}

func NewRaftServer(n *Node) *RaftServer {
	return &RaftServer{node: n}
}

func (s *RaftServer) RequestVote(ctx context.Context, args *pb.RequestVoteArgs) (*pb.RequestVoteReply, error) {
	reply := s.node.HandleRequestVote(RequestVoteArgs{
		Term:         int(args.Term),
		CandidateId:  args.CandidateId,
		LastLogIndex: int(args.LastLogIndex),
		LastLogTerm:  int(args.LastLogTerm),
	})

	return &pb.RequestVoteReply{
		Term:        int32(reply.Term),
		VoteGranted: reply.VoteGranted,
	}, nil
}

func (s *RaftServer) AppendEntries(ctx context.Context, args *pb.AppendEntriesArgs) (*pb.AppendEntriesReply, error) {
	entries := make([]LogEntry, len(args.Entries))
	for i, e := range args.Entries {
		entries[i] = LogEntry{Term: int(e.Term), Command: Command{Key: e.Key, Value: e.Value}}
	}

	reply := s.node.HandleAppendEntries(AppendEntriesArgs{
		Term:         int(args.Term),
		LeaderId:     args.LeaderId,
		PrevLogIndex: int(args.PrevLogIndex),
		PrevLogTerm:  int(args.PrevLogTerm),
		Entries:      entries,
		LeaderCommit: int(args.LeaderCommit),
	})

	return &pb.AppendEntriesReply{
		Term:    int32(reply.Term),
		Success: reply.Success,
	}, nil
}
