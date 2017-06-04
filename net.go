package raft

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
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
	Term        int
	CandidateID string
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
	Self         *heartBeat
	LastLogIndex int
	LastLogTerm  int
}

type voteResponse struct {
	Term int
	Vote *Vote
}

// ListenTCP listens for the RPC calls
func (r *Raft) ListenTCP() {
	return
	// rpcs := rpc.NewServer()
	// rpc.Register(r)
	// for {
	// 	conn, err := r.tcpListener.Accept()
	// 	if err != nil {
	// 		log.Printf("Failed to accept. Err: %s", err)
	// 		continue
	// 	}
	//
	// 	fmt.Printf("Accepted the connection from %s\n", conn.RemoteAddr())
	//
	// 	go rpcs.ServeConn(conn)
	//
	// }
}

var start time.Time

// ListenUDP listens for the udp packages
func (r *Raft) ListenUDP() {
	buffer := make([]byte, udpBufferSize)

	var addr net.Addr
	var err error

	for {
		buf := buffer[0:udpBufferSize]

		_, addr, err = r.udpListener.ReadFrom(buf)

		if err != nil {
			log.Printf("Failed to accept UDP. Err: %s", err)
			continue
		}

		msgType := binary.BigEndian.Uint16(buf[:2])
		buf = buf[2:]

		switch msgType {
		case heartBeatMsg:
			// fmt.Printf("Got the Heart beat from <--> %s\n", addr)
			r.handleHeartBeat(addr, buf)
		case voteRequestMsg:
			fmt.Printf("Got the Vote Request from <--> %s\n", addr)
			r.handleVoteRequest(addr, buf)
		case voteResponseMsg:
			// fmt.Printf("Got the Heart beat Response from <--> %s\n", addr)
			r.handleVoteResponse(addr, buf)
		case joinPingMsg:
			// fmt.Printf("Got the Join Ping from <--> %s\n", addr)
			go r.handleJoinPing(addr, buf)
		case joinAckRespMsg:
			// fmt.Printf("Got the Join Ack Resp from <--> %s\n", addr)
			r.handleJoinAckResp(addr, buf)
		default:
			fmt.Printf("couldn't parse the msg from %s", addr)
			continue
		}

	}
}

func (r *Raft) handleJoinPing(from net.Addr, msg []byte) {
	var p *joinPing

	err := decode(msg, &p)
	if err != nil {
		log.Printf("Failed to read the ping UDP packet. Err %s", err)
		return
	}

	// add to netradler nodes
	err = r.addNode(p.Node)
	if err != nil {
		log.Printf("Failed to add the node information. Err %s", err)
		return
	}

	err = r.invokeJoinAckResp(from, p)
	if err != nil {
		log.Printf("Failed to invoke JoinAckResp. Err %s", err)
	}
}

func (r *Raft) handleJoinAckResp(from net.Addr, msg []byte) {
	var ack joinAckResp

	err := decode(msg, &ack)
	if err != nil {
		log.Printf("Failed to read the ack response UDP Packet. Err %s", err)
	}

	err = r.syncStateAfterJoin(&ack)
	if err != nil {
		log.Printf("Failed to sync state after join. Err %s", err)
	}

}

func (r *Raft) handleHeartBeat(from net.Addr, msg []byte) {
	start = time.Now()
	var hb heartBeat

	err := decode(msg, &hb)
	if err != nil {
		log.Printf("Failed to read the HeartBeat UDP packet. Err %s", err)
		return
	}

	if len(r.Nodes()) <= 0 && r.self.state != Unknown {
		go r.joinToClusterByAddr(from)
		// Altay return to here and finish this workflow...
		return
	}
	r.invokeVoteReponse(from, &hb)
	fmt.Println("HandleHeartBeat: ", time.Since(start))
}

func (r *Raft) handleVoteRequest(from net.Addr, msg []byte) {
	vr := &voteRequest{}

	err := decode(msg, &vr)
	if err != nil {
		log.Printf("Failed to read the HeartBeat UDP packet. Err %s", err)
		return
	}

	if len(r.Nodes()) <= 0 {
		return
	}

	r.invokeVoteReponse(from, vr.Self)
}

func (r *Raft) invokeVoteReponse(from net.Addr, hb *heartBeat) {
	vr := &voteResponse{}
	v := &Vote{}

	v.Voter = r.config.Name

	if hb.Term >= r.self.term {
		r.hbChan <- hb

		r.stateLock.Lock()
		r.self.term = hb.Term
		r.stateLock.Unlock()

		v.VoteStatus = Voted

	} else {
		vr.Term = r.self.term
		v.VoteStatus = NotVoted
	}

	vr.Vote = v

	err := r.encodeAndSendMsg(from, voteResponseMsg, &vr)
	if err != nil {
		log.Printf("Failed to send the reponse to the heart beat. Err %s", err)
	}
}

func (r *Raft) handleVoteResponse(addr net.Addr, msg []byte) {
	vr := &voteResponse{}

	err := decode(msg, &vr)
	if err != nil {
		log.Printf("Failed to read the HeartBeat UDP packet. Err %s", err)
		return
	}

	if r.self.state != Leader {
		r.votes <- vr.Vote
	}

	if vr.Term > r.self.term {
		r.self.term = vr.Term
	}

}
