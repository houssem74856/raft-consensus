# Raft Consensus

Implementation of the [Raft distributed consensus algorithm](https://raft.github.io/raft.pdf) in Go. Raft is an algorithm applied to distributed databases where it ensures consistency between nodes and also can
survive node failures and network partitions as long as the majority are alive and can communicate. In this implementation each node runs in its own docker container and communicates with other nodes 
using gRPC.

## Why Raft

Raft sits on the CP side of the CAP theorem where if a partition happens it prioritizes consistency over availability. This makes it a good fit for systems where correctness is critical but throughput is not the 
main concern because guess what the consistency part takes time to ensure, so raft won't be as fast as something like cassandra but you can ensure that data inconsistencies and conflicts won't happen. Raft is 
used in production by some well known systems such as etcd: a distributed key-value store used for Kubernetes cluster state management, CockroachDB: a resilient and consistent distributed relational database, 
and Consul: platform for service discovery and configuration management.

## Raft Core Componenets

### leader election

Every node starts as a follower, if a follower doesn't hear from a leader within a randomized timeout, it transitions to candidate, votes for itself, and requests votes from other nodes. To become leader a node must get
the majority of votes. A follower node gives its vote to the first candidate that asks as long the candidate's term is greater or equal to its own. The randomized timeout is a smart trick to minimize the chance of split votes 
happening because if it was the same timeout for all nodes, then they would be much more likely to become candidates at the same time and split the votes without anyone reaching majority.

### log replication

When a node becomes a leader it starts sending periodic heartbeats to other nodes to say I am still alive. These heartbeats also carry log entries, so followers can catch up to the leader's latest state. When a client submits a 
key-value entry, the leader adds it to its log, and sends it to all nodes, and when the majority adds it too, the leader commits it. A committed entry is guaranteed to survive any future leader failure.

## Implementation

### gRPC

For communication between nodes gRPC was used. gRPC keeps persistent connections open and uses binary serialization, making it faster and more efficient for high frequency internal communication like heartbeats and 
vote requests.

### Docker

each node is put in its own container, the idea behind this is to make simulating node failures easier, which is pretty important to test if the implementation is correct and functions as it should.

### Go

for efficiency, heartbeats and vote requests are sent to all peers simultaneously using goroutines, instead of waiting for each reply before contacting the next node. Channels were also a good fit for coordination
tasks between goroutines like collecting votes and signaling an election timer reset.
