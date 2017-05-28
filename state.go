package raft

import "net"

// States of node
const (
	Follower = iota
	Candidate
	Leader
)

// State ...
type State struct {
	state    int
	term     int
	leaderID string
	vote     int
	votedFor string
}

// Node ...
type Node struct {
	Name string
	Addr *net.UDPAddr
}
