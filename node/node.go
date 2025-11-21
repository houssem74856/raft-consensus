package node

import (
	"fmt"
	"math/rand"
	"time"
)

type Node struct {
	Id                string
	Peers             []string
	State             *State
	RPC               RPC
	ElectionResetChan chan Empty
}

type Empty struct{}

func NewNode(id string, peers []string, rpc RPC) *Node {
	return &Node{
		Id:                id,
		Peers:             peers,
		State:             NewState(),
		RPC:               rpc,
		ElectionResetChan: make(chan Empty, 1),
	}
}

func (n *Node) Start() {
	go n.runElectionTimer()
}

func (n *Node) runElectionTimer() {
	for {
		timeout := time.Duration(150+rand.Intn(150)) * time.Millisecond
		timer := time.NewTimer(timeout)

		select {
		case <-timer.C:
			n.State.mu.Lock()
			isFollower := n.State.Role == Follower
			n.State.mu.Unlock()

			if isFollower /*&& (n.Id == "n1" || n.Id == "n2")*/ /*split vote simulation for four nodes*/ {
				n.startElection()
			}
		case <-n.ElectionResetChan:
			timer.Stop()
		}
	}
}

func (n *Node) startElection() {
	n.State.mu.Lock()
	n.State.CurrentTerm++
	n.State.Role = Candidate
	n.State.VotedFor = n.Id
	term := n.State.CurrentTerm
	n.State.mu.Unlock()

	votes := 1
	majority := ((len(n.Peers) + 1) / 2) + 1

	votesChan := make(chan RequestVoteReply, len(n.Peers))

	for _, peer := range n.Peers {
		p := peer
		go func() {
			reply := n.RPC.SendRequestVote(p, RequestVoteArgs{
				Term:        term,
				CandidateId: n.Id,
			})

			votesChan <- reply
		}()
	}

	electionTimeout := time.After(150 * time.Millisecond)

	for i := 0; i < len(n.Peers); i++ {
		select {
		case reply := <-votesChan:
			if reply.Term > term {
				n.State.mu.Lock()
				n.State.CurrentTerm = reply.Term
				n.State.Role = Follower
				n.State.VotedFor = ""
				n.State.mu.Unlock()
				return
			}

			if reply.VoteGranted {
				votes++

				if votes >= majority {
					n.State.mu.Lock()
					n.State.Role = Leader
					n.State.mu.Unlock()

					fmt.Println("Leader elected:", n.Id, "term:", term)

					n.startHeartbeat()
					return
				}
			}

		case <-electionTimeout:
			fmt.Println("Election timed out for", n.Id)
			n.State.mu.Lock()
			n.State.Role = Follower
			n.State.VotedFor = ""
			n.State.mu.Unlock()
			return
		}
	}

	n.State.mu.Lock()
	n.State.Role = Follower
	n.State.VotedFor = ""
	n.State.mu.Unlock()
}

func (n *Node) startHeartbeat() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		n.State.mu.Lock()
		if n.State.Role != Leader {
			n.State.mu.Unlock()
			fmt.Println(n.Id, "not leader anymore")
			return
		}
		peers := n.Peers
		term := n.State.CurrentTerm
		n.State.mu.Unlock()

		for _, peer := range peers {
			p := peer
			go func() {
				reply := n.RPC.SendAppendEntries(p, AppendEntriesArgs{
					Term:     term,
					LeaderId: n.Id,
				})

				if reply.Term > term {
					n.State.mu.Lock()
					n.State.CurrentTerm = reply.Term
					n.State.Role = Follower
					n.State.VotedFor = ""
					n.State.mu.Unlock()
				}
			}()
		}

		<-ticker.C
	}
}

func (n *Node) HandleRequestVote(args RequestVoteArgs) RequestVoteReply {
	n.State.mu.Lock()
	defer n.State.mu.Unlock()

	// split vote simulation for four nodes
	/*if args.Term == 1 {
		if (args.CandidateId == "n1" && n.Id == "n3") || (args.CandidateId == "n2" && n.Id == "n4") {
			n.State.CurrentTerm = args.Term
			n.State.Role = Follower
			n.State.VotedFor = args.CandidateId
			return RequestVoteReply{Term: n.State.CurrentTerm, VoteGranted: true}
		} else {
			return RequestVoteReply{Term: n.State.CurrentTerm, VoteGranted: false}
		}
	}*/

	if args.Term > n.State.CurrentTerm {
		n.State.CurrentTerm = args.Term
		n.State.Role = Follower
		n.State.VotedFor = ""
	}

	if args.Term >= n.State.CurrentTerm && n.State.VotedFor == "" {
		select {
		case n.ElectionResetChan <- Empty{}:
		default:
		}

		n.State.VotedFor = args.CandidateId
		return RequestVoteReply{Term: n.State.CurrentTerm, VoteGranted: true}
	}

	return RequestVoteReply{Term: n.State.CurrentTerm, VoteGranted: false}
}

func (n *Node) HandleAppendEntries(args AppendEntriesArgs) AppendEntriesReply {
	n.State.mu.Lock()
	defer n.State.mu.Unlock()

	if args.Term > n.State.CurrentTerm {
		n.State.CurrentTerm = args.Term
		n.State.Role = Follower
		n.State.VotedFor = ""
	}

	if args.Term >= n.State.CurrentTerm {
		select {
		case n.ElectionResetChan <- Empty{}:
		default:
		}

		return AppendEntriesReply{Term: n.State.CurrentTerm, Success: true}
	}

	return AppendEntriesReply{Term: n.State.CurrentTerm, Success: false}
}
