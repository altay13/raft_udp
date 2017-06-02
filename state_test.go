package raft

import (
	"testing"
	"time"
)

func TestState_handleState(t *testing.T) {
	r := GetRaft(t)

	r.stateCh <- Candidate
	if r.self.state != Candidate {
		t.Fatalf("Failed to change the state to Candidate. %d", r.self.state)
	}

	r.stateCh <- Leader
	if r.self.state != Leader {
		t.Fatalf("Failed to change the state to Leader. %d", r.self.state)
	}

	r.stateCh <- Follower
	if r.self.state != Follower {
		t.Fatalf("Failed to change the state to Follower. %d", r.self.state)
	}
}

func TestState_voting(t *testing.T) {
	r := GetRaft(t)

	r.stateCh <- Candidate
	time.Sleep(2 * time.Second)

	if r.self.state != Leader {
		t.Fatalf("Failed to acquire leader after the candidate state change when alone. State is: %d", r.self.state)
	}

	r.stateCh <- Leader

	r2 := JoinRaft(t, "Machine 2", r.config.BindUDPPort)
	time.Sleep(2 * time.Second)

	if r2.self.leaderID != r.config.Name {
		t.Fatalf("Failed to acquire the leadership. Node: Should be `Machine 1` but instead: %s", r2.self.leaderID)
	}

	r3 := JoinRaft(t, "Machine 3", r.config.BindUDPPort)
	time.Sleep(2 * time.Second)

	if r3.self.leaderID != r.config.Name {
		t.Fatalf("Failed to acquire the leadership. Node: Should be `Machine 1` but instead: %s", r3.self.leaderID)
	}

	r3.stateCh <- Candidate
	time.Sleep(2 * time.Second)

	if r3.self.state != Leader {
		t.Fatalf("Failed to achieve the leadership for Machine 3. But instead: %s", r3.self.leaderID)
	}

}

func TestState_listenHeartBeat(t *testing.T) {
	r := GetRaft(t)
	r.stateCh <- Leader

	r2 := JoinRaft(t, "Machine 2", r.config.BindUDPPort)
	time.Sleep(2 * time.Second)

	r.Shutdown()

	time.Sleep(1 * time.Second)

	if r2.self.state != Leader {
		t.Fatalf("Failed to achieve the leadership for Machine 2. Leader is %s", r2.self.leaderID)
	}
}

func TestState_becomeLeader(t *testing.T) {
	r := GetRaft(t)
	r.becomeLeader()
	time.Sleep(100 * time.Millisecond)

	if r.self.state != Leader {
		t.Fatalf("Failed to become a leader")
	}
}

//
// func TestState_becomeCandidate(t *testing.T) {
// 	r := GetRaft(t)
// 	r.becomeCandidate()
// 	time.Sleep(20 * time.Millisecond)
//
// 	if r.self.state != Candidate {
// 		t.Fatalf("Failed to become a candidate")
// 	}
// }
//
// func TestState_becomeFollower(t *testing.T) {
// 	r := GetRaft(t)
// 	r.becomeFollower()
//
// 	if r.self.state != Follower {
// 		t.Fatalf("Failed to become a follower")
// 	}
// }

func TestState_broadcastVoteCollection(t *testing.T) {
	// r := GetRaft(t)
	// Will add this later...
}
