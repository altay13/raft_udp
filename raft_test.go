package raft

import (
	"testing"
	"time"
)

func GetRaft(t *testing.T) *Raft {
	conf := DefaultConf()
	r, err := Init(conf)

	if err != nil {
		t.Fatalf("Couldn't create the netradler")
	}

	return r
}

func TestRaft_CreateShutdown(t *testing.T) {
	r := GetRaft(t)
	defer r.Shutdown()

	if r.self.state != Follower {
		t.Fatalf("On init the state should be 'Follower' = 0, instead: %d", r.self.state)
	}

	if r.self.term != 0 {
		t.Fatalf("On init the term state should be 0, instead: %d", r.self.term)
	}

	time.Sleep(time.Duration(r.config.ElectionTime)*time.Second + 100*time.Millisecond)

	if r.self.state != Candidate {
		t.Fatalf("Election timeout fired, the state should be 'Candidate' = 1, instead: %d", r.self.state)
	}

}
