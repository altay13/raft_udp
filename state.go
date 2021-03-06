package raft

import (
	"fmt"
	"net"
	"time"
)

// States of node
const (
	Follower = iota
	Candidate
	Leader
	Unknown
)

// Voted ...
const (
	Voted = iota
	NotVoted
	UnknownVote
)

// State ...
type State struct {
	state    int
	term     int
	leaderID string
	vote     *Vote
}

// Node ...
type Node struct {
	Name       string
	Addr       *net.UDPAddr
	VoteStatus int
}

// NodeJoin ...
type NodeJoin struct {
	Name string
	Addr *net.UDPAddr
}

// Vote ...
type Vote struct {
	Voter      string
	VoteStatus int
}

func (r *Raft) handleState() {
	for {
		select {
		case s := <-r.stateCh:
			switch s {
			case Follower:
				fmt.Println("State Changed to: Follower")

				r.stopHeartBeatTicker()
				r.stopVotingTicker()

				r.stateLock.Lock()
				r.self.state = Follower
				r.stateLock.Unlock()

				go r.listenHeartBeat()
			case Candidate:
				fmt.Println("State Changed to: Candidate")

				r.releaseVotes()

				r.stateLock.Lock()
				r.self.state = Candidate
				r.self.term++
				r.stateLock.Unlock()

				r.stopHeartBeatTicker()

				// r.voteForSelf()
				r.broadcastVoteCollection()
				go r.scheduleVoting()
			case Leader:
				fmt.Println("State Changed to: Leader")

				r.stateLock.Lock()
				r.self.state = Leader
				r.self.leaderID = r.config.Name
				r.stateLock.Unlock()

				go func() {
					r.scheduleHeartBeat()
					r.stopTicker()
				}()

				r.releaseVotes()
			case Unknown:
				go func() {
					r.stopHeartBeatTicker()
					r.stopTicker()
					r.stopVotingTicker()
				}()
			}
		default:
			continue
		}
	}
}

func (r *Raft) scheduleVoting() {
	r.resetVotingTicker()
	for {
		select {
		case vt := <-r.votes:
			fmt.Println("vote: ", vt)
			r.addVote(vt)
		case <-r.votingTimeoutTicker.C:
			// check the votes and decide what to do next --> Candidate again or Leader
			fmt.Println("Voting timer ticked: ", r.self.term, "time:", time.Now())
			r.stopVotingTicker()
			if r.calculateVotes() {
				r.becomeLeader()
			} else {
				r.becomeCandidate()
			}
			return
		default:
			continue
		}
	}
}

func (r *Raft) listenHeartBeat() {
	r.resetTicker()
	go func() {
		for {
			select {
			case hb := <-r.hbChan:
				fmt.Println("Got the HB", hb)
				r.self.leaderID = hb.CandidateID

				if r.self.state != Follower {
					r.becomeFollower()
					return
				} else {
					r.resetTicker()
				}
			case <-r.electionTimeoutTicker.C:
				// r.stopTicker() can not stop hb listen, because any time the leader hb can come...
				// Then I have to compare the terms and decide what to do next
				r.stopTicker()
				fmt.Println("Election timeout fired, have to change the state...")
				r.becomeCandidate()
			default:
				continue
			}
		}
	}()
}

func (r *Raft) becomeLeader() {
	r.stateCh <- Leader
}

func (r *Raft) becomeCandidate() {
	r.stateCh <- Candidate
}

func (r *Raft) becomeFollower() {
	r.stateCh <- Follower
}

func (r *Raft) broadcastVoteCollection() error {
	vr := &voteRequest{
		Self: &heartBeat{
			CandidateID: r.config.Name,
			Term:        r.self.term,
		},
	}

	fmt.Println("Term: ", vr.Self.Term)

	for _, node := range r.nodes {
		go r.encodeAndSendMsg(node.Addr, voteRequestMsg, &vr)
	}

	fmt.Println("The vote collection broadcasted")

	return nil
}

func (r *Raft) broadcastHeartBeat() error {

	hb := &heartBeat{}
	hb.Term = r.self.term
	hb.CandidateID = r.config.Name

	for _, node := range r.nodes {
		go r.encodeAndSendMsg(node.Addr, heartBeatMsg, &hb)
	}

	return nil
}

func (r *Raft) scheduleHeartBeat() {
	r.resetHeartBeat()

	go func() {
		for {
			select {
			case <-r.heartBeat.C:
				r.broadcastHeartBeat()
			case <-r.stopHeartBeat:
				return
			}
		}
	}()
}
