package raft

import (
	"testing"
	"time"
)

func GetRaft(t *testing.T) *Raft {
	conf := DefaultConf()

	conf.Name = "Machine 1"
	conf.ElectionTime = 300

	for i := 0; i < 100; i++ {
		r, err := Init(conf)
		if err == nil {
			return r
		}
		conf.BindTCPPort++
		conf.BindUDPPort++
	}

	t.Fatalf("Couldn't create the raft node")
	return nil
}

func JoinRaft(t *testing.T, name string, port int) *Raft {
	conf := DefaultConf()

	conf.Name = name
	conf.ElectionTime = 300

	conf.BindTCPPort = 2222
	conf.BindUDPPort = 2222

	for i := 0; i < 100; i++ {
		r, err := Join(conf, "127.0.0.1", port)
		if err == nil {
			return r
		}
		conf.BindTCPPort++
		conf.BindUDPPort++
	}

	t.Fatalf("Couldn't create the raft node")
	return nil
}

func TestRaft_CreateShutdown(t *testing.T) {
	r := GetRaft(t)
	// defer r.Shutdown()

	if r.self.state != Follower {
		t.Fatalf("On init the state should be 'Follower' = 0, instead: %d", r.self.state)
	}

	if r.self.term != 0 {
		t.Fatalf("On init the term state should be 0, instead: %d", r.self.term)
	}

	time.Sleep(time.Duration(r.config.ElectionTime)*time.Millisecond + 50*time.Millisecond)

	if r.self.state != Candidate {
		t.Fatalf("Election timeout fired, the state should be 'Candidate' = 1, instead: %d", r.self.state)
	}
}

func TestRaft_Join(t *testing.T) {
	r1 := GetRaft(t)
	// defer r1.Shutdown()

	r2 := JoinRaft(t, "Machine 2", r1.config.BindUDPPort)
	// defer r2.Shutdown()

	t.Log("Successfully created the instances")

	time.Sleep(2000 * time.Millisecond)

	node1 := r1.nodeMap[r2.config.Name]
	node2 := r2.nodeMap[r1.config.Name]

	if node1.Name != r2.config.Name {
		t.Fatalf("Failed to join nodes")
	}

	if node2.Name != r1.config.Name {
		t.Fatalf("Failed to join nodes")
	}

	if r2.self.leaderID != r1.config.Name {
		t.Fatalf("Failed to achieve the leadership: %s vs %s", r2.self.leaderID, r1.config.Name)
	}
}

func TestRaft_Nodes(t *testing.T) {
	r := GetRaft(t)

	n1 := &Node{
		Addr:       nil,
		Name:       "Machine 2",
		VoteStatus: Unknown,
	}
	n2 := &Node{
		Addr:       nil,
		Name:       "Machine 3",
		VoteStatus: Unknown,
	}
	n3 := &Node{
		Addr:       nil,
		Name:       "Machine 4",
		VoteStatus: Unknown,
	}

	r.addNode(n1)
	r.addNode(n2)
	r.addNode(n3)

	if len(r.Nodes()) != 3 {
		t.Fatalf("The length of Nodes should be 3, but instead: %d", len(r.Nodes()))
	}
}
