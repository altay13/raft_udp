package raft

import (
	"fmt"
	"net"
)

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

func (r *Raft) handleState() {
	for {
		select {
		case s := <-r.stateCh:
			switch s {
			case Follower:
				fmt.Println("State Changed to: Follower")
				break
			case Candidate:
				fmt.Println("State Changed to: Candidate")
				break
			case Leader:
				fmt.Println("State Changed to: Leader")
				break
			}
		default:
			continue
		}
	}
}
