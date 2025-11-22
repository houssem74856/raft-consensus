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
	SM                *StateMachine

	nextIndex  map[string]int
	matchIndex map[string]int
}

type Empty struct{}

func NewNode(id string, peers []string, rpc RPC) *Node {
	return &Node{
		Id:                id,
		Peers:             peers,
		State:             NewState(),
		RPC:               rpc,
		ElectionResetChan: make(chan Empty, 1),
		SM:                NewStateMachine(),
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

	lastIndex := len(n.State.Log) - 1
	lastTerm := 0
	if lastIndex >= 0 {
		lastTerm = n.State.Log[lastIndex].Term
	}
	n.State.mu.Unlock()

	votes := 1
	majority := ((len(n.Peers) + 1) / 2) + 1

	votesChan := make(chan RequestVoteReply, len(n.Peers))

	for _, peer := range n.Peers {
		p := peer
		go func() {
			reply := n.RPC.SendRequestVote(p, RequestVoteArgs{
				Term:         term,
				CandidateId:  n.Id,
				LastLogIndex: lastIndex,
				LastLogTerm:  lastTerm,
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
					lastIndex := len(n.State.Log)
					n.nextIndex = map[string]int{}
					n.matchIndex = map[string]int{}
					for _, p := range n.Peers {
						n.nextIndex[p] = lastIndex
						n.matchIndex[p] = 0
					}
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
		for _, peer := range n.Peers {
			p := peer
			go func() {
				n.State.mu.Lock()
				ni := n.nextIndex[p]
				prevLogIndex := ni - 1
				prevLogTerm := 0
				if prevLogIndex >= 0 {
					prevLogTerm = n.State.Log[prevLogIndex].Term
				}
				entries := []LogEntry{}
				if len(n.State.Log) > 0 {
					entries = append([]LogEntry{}, n.State.Log[ni:]...)
				}
				term := n.State.CurrentTerm
				leaderCommit := n.State.CommitIndex
				n.State.mu.Unlock()

				reply := n.RPC.SendAppendEntries(p, AppendEntriesArgs{
					Term:         term,
					LeaderId:     n.Id,
					PrevLogIndex: prevLogIndex,
					PrevLogTerm:  prevLogTerm,
					Entries:      entries,
					LeaderCommit: leaderCommit,
				})

				n.handleAppendEntriesReply(p, reply, len(entries))
			}()
		}

		<-ticker.C

		n.State.mu.Lock()
		role := n.State.Role
		n.State.mu.Unlock()

		if role != Leader {
			fmt.Println(n.Id, "is not leader anymore")
			return
		}
	}
}

func (n *Node) handleAppendEntriesReply(peer string, reply AppendEntriesReply, sentEntries int) {
	n.State.mu.Lock()
	defer n.State.mu.Unlock()

	if reply.Term > n.State.CurrentTerm {
		n.State.CurrentTerm = reply.Term
		n.State.Role = Follower
		n.State.VotedFor = ""

		return
	}

	if reply.Success {
		n.matchIndex[peer] += sentEntries
		n.nextIndex[peer] = n.matchIndex[peer]
		n.updateCommitIndex()
		return
	}

	n.nextIndex[peer]--
}

func (n *Node) updateCommitIndex() {
	for i := len(n.State.Log) - 1; i > n.State.CommitIndex; i-- {
		count := 1
		for _, peer := range n.Peers {
			if n.matchIndex[peer] >= i {
				count++
			}
		}

		if count >= ((len(n.Peers)+1)/2)+1 && n.State.Log[i].Term == n.State.CurrentTerm {
			n.State.CommitIndex = i
			n.applyEntries()
			return
		}
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

	lastIndex := len(n.State.Log) - 1
	lastTerm := 0
	if lastIndex >= 0 {
		lastTerm = n.State.Log[lastIndex].Term
	}

	upToDate := false
	if args.LastLogTerm > lastTerm {
		upToDate = true
	} else if args.LastLogTerm == lastTerm && args.LastLogIndex >= lastIndex {
		upToDate = true
	}

	if args.Term >= n.State.CurrentTerm && n.State.VotedFor == "" && upToDate {
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

	if args.Term < n.State.CurrentTerm {
		return AppendEntriesReply{Term: n.State.CurrentTerm, Success: false}
	}

	select {
	case n.ElectionResetChan <- Empty{}:
	default:
	}

	if args.Term > n.State.CurrentTerm {
		n.State.CurrentTerm = args.Term
		n.State.Role = Follower
		n.State.VotedFor = ""
	}

	// log consistency check
	if args.PrevLogIndex >= 0 && (args.PrevLogIndex >= len(n.State.Log) || n.State.Log[args.PrevLogIndex].Term != args.PrevLogTerm) {
		/*// conflict optimization (optional)
		conflictIndex := len(n.State.Log)
		if args.PrevLogIndex < len(n.State.Log) {
			term := n.State.Log[args.PrevLogIndex].Term
			for i := args.PrevLogIndex; i >= 0 && n.State.Log[i].Term == term; i-- {
				conflictIndex = i
			}
		}*/

		return AppendEntriesReply{
			Term:    n.State.CurrentTerm,
			Success: false,
		}
	}

	// append new entries
	i := 0
	for ; i < len(args.Entries); i++ {
		pos := args.PrevLogIndex + 1 + i
		if pos < len(n.State.Log) {
			if n.State.Log[pos].Term != args.Entries[i].Term {
				n.State.Log = n.State.Log[:pos]
				break
			}
		} else {
			break
		}
	}

	n.State.Log = append(n.State.Log, args.Entries[i:]...)

	// commit index update
	if args.LeaderCommit > n.State.CommitIndex {
		//fmt.Println("yeeees")
		n.State.CommitIndex = min(args.LeaderCommit, len(n.State.Log)-1)
		n.applyEntries()
	}

	return AppendEntriesReply{Term: n.State.CurrentTerm, Success: true}
}

func (n *Node) applyEntries() {
	for n.State.LastApplied < n.State.CommitIndex {
		n.State.LastApplied++
		entry := n.State.Log[n.State.LastApplied]

		n.SM.Apply(entry)
	}
}

/*func (n *Node) SubmitCommand(cmd Command) (ok bool, term int) {
	n.State.mu.Lock()
	defer n.State.mu.Unlock()

	if n.State.Role != Leader {
		return false, n.State.CurrentTerm
	}

	entry := LogEntry{
		Term:    n.State.CurrentTerm,
		Command: cmd,
	}
	n.State.Log = append(n.State.Log, entry)

	return true, n.State.CurrentTerm
}*/
