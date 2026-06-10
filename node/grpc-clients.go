package node

import (
	"context"
	"fmt"
	pb "raft-consensus/proto"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClients struct {
	clients map[string]pb.RaftClient
}

func NewGrpcClients(peers map[string]string) *GrpcClients {
	clients := make(map[string]pb.RaftClient)

	for id, addr := range peers {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			fmt.Println("failed to connect to", addr, ":", err)
		}

		clients[id] = pb.NewRaftClient(conn)
	}

	return &GrpcClients{clients: clients}
}

func (gcs *GrpcClients) SendRequestVote(to string, args RequestVoteArgs) RequestVoteReply {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	reply, err := gcs.clients[to].RequestVote(ctx, &pb.RequestVoteArgs{
		Term:         int32(args.Term),
		CandidateId:  args.CandidateId,
		LastLogIndex: int32(args.LastLogIndex),
		LastLogTerm:  int32(args.LastLogTerm),
	})

	if err != nil {
		return RequestVoteReply{}
	}

	return RequestVoteReply{Term: int(reply.Term), VoteGranted: reply.VoteGranted}
}

func (gcs *GrpcClients) SendAppendEntries(to string, args AppendEntriesArgs) AppendEntriesReply {
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	entries := make([]*pb.LogEntry, len(args.Entries))
	for i, e := range args.Entries {
		entries[i] = &pb.LogEntry{Term: int32(e.Term), Key: e.Command.Key, Value: e.Command.Value}
	}

	reply, err := gcs.clients[to].AppendEntries(ctx, &pb.AppendEntriesArgs{
		Term:         int32(args.Term),
		LeaderId:     args.LeaderId,
		PrevLogIndex: int32(args.PrevLogIndex),
		PrevLogTerm:  int32(args.PrevLogTerm),
		Entries:      entries,
		LeaderCommit: int32(args.LeaderCommit),
	})

	if err != nil {
		return AppendEntriesReply{}
	}

	return AppendEntriesReply{Term: int(reply.Term), Success: reply.Success}
}
