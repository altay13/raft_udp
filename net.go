package raft

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/rpc"
)

const (
	udpBufferSize = 65536
)

const (
	heartBeatMsg = iota
	joinPingMsg
	joinAckRespMsg
	voteRequestMsg
	voteResponseMsg
)

type msgHeader struct {
	MsgType int
}

type heartBeat struct {
	Term int
	Name string
}

type joinPing struct {
	Node *Node
}

type joinAckResp struct {
	LeaderID string
	Nodes    []*Node
	Node     *Node
}

type voteRequest struct {
	Term         int
	CandidateID  string
	LastLogIndex int
	LastLogTerm  int
}

type voteResponse struct {
	Term        int
	VoteGranted bool
}

// ListenTCP listens for the RPC calls
func (n *Raft) ListenTCP() {
	rpcs := rpc.NewServer()
	rpc.Register(n)
	for {
		conn, err := n.tcpListener.Accept()
		if err != nil {
			log.Printf("Failed to accept. Err: %s", err)
			continue
		}

		fmt.Printf("Accepted the connection from %s\n", conn.RemoteAddr())

		go rpcs.ServeConn(conn)

	}
}

// ListenUDP listens for the udp packages
func (n *Raft) ListenUDP() {
	buffer := make([]byte, udpBufferSize)

	var addr net.Addr
	var err error

	for {
		buf := buffer[0:udpBufferSize]

		_, addr, err = n.udpListener.ReadFrom(buf)

		if err != nil {
			log.Printf("Failed to accept UDP. Err: %s", err)
			continue
		}

		msgType := binary.BigEndian.Uint16(buf[:2])
		buf = buf[2:]

		switch msgType {
		// case heartBeatMsg:
		// 	fmt.Printf("Got the Ping from Leader <--> %s\n", addr)
		// 	n.handleLeaderPing(addr, buf)
		// case pingMsg:
		// 	fmt.Printf("Got the Ping from %s\n", addr)
		// 	n.handlePing(addr, buf)
		// case ackRespMsg:
		// 	fmt.Printf("Got the AckRespPing from %s\n", addr)
		// 	n.handleAck(addr, buf)
		// case joinPingMsg:
		// 	fmt.Printf("Got the joinPing from %s\n", addr)
		// 	n.handleJoinPing(addr, buf)
		// case joinAckRespMsg:
		// 	fmt.Printf("Got the joinAckResp from %s\n", addr)
		// 	n.handleJoinAckResp(addr, buf)
		// case voteRequestMsg:
		// 	fmt.Printf("Got the vote request from %s\n", addr)
		// 	n.handleVoteRequest(addr, buf)
		// case voteResponseMsg:
		// 	fmt.Printf("Got the vote responce from %s\n", addr)
		// 	n.handleVoteResponse(addr, buf)
		default:
			fmt.Printf("couldn't parse the msg from %s", addr)
			continue
		}

	}
}
