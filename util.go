package raft

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

func decode(buf []byte, out interface{}) error {

	// create the buffer from input buf []bytes
	msgBuff := bytes.NewBuffer(buf)

	// create the decoder, which will decode the r buf
	dec := gob.NewDecoder(msgBuff)

	// will write the decoded value to out. out can be any struct after the decode (if compatible)
	return dec.Decode(out)
}

func encode(msgType int, msg interface{}) (*bytes.Buffer, error) {

	// Create the buffer
	buf := bytes.NewBuffer(nil)

	// add the buffer type, do not encode it
	binary.Write(buf, binary.BigEndian, uint16(msgType))

	// create encoder. This enc will encode to buf
	enc := gob.NewEncoder(buf)

	// encode the message in itself to buf
	err := enc.Encode(msg)

	return buf, err
}

func getUDPAddr(addr string, port int) *net.UDPAddr {
	res := &net.UDPAddr{IP: net.ParseIP(addr), Port: port}

	return res
}

func (r *Raft) getTCPAddress() (addr string) {
	addr = fmt.Sprintf("%s:%d", r.config.BindAddr, r.config.BindTCPPort)

	return addr
}

func (r *Raft) getUDPAddress() (addr string) {
	addr = fmt.Sprintf("%s:%d", r.config.BindAddr, r.config.BindUDPPort)

	return addr
}

func getElectionTicker(n int) *time.Ticker {
	rnd := int(rand.Float32() * 1000)
	intrv := n*1000 + rnd

	return time.NewTicker(time.Duration(intrv) * time.Millisecond)
}

func (r *Raft) encodeAndSendMsg(dest net.Addr, msgType int, msg interface{}) error {
	buf, err := encode(msgType, msg)
	if err != nil {
		return err
	}

	if err := r.sendMsg(dest, buf); err != nil {
		return err
	}

	return nil
}

func (r *Raft) sendMsg(to net.Addr, msg *bytes.Buffer) error {
	_, err := r.udpListener.WriteTo(msg.Bytes(), to)
	if err != nil {
		log.Printf("Failed to send UDP to %s, Err: %s", to, err)
		return err
	}
	return nil
}

func (r *Raft) voteForSelf() {
	vt := &Vote{
		votedFor:    r.config.Name,
		voteGranted: true,
	}
	r.votes <- vt
}

func (r *Raft) addVote(vt *Vote) {
	node := r.nodeMap[vt.votedFor]
	if node == nil {
		log.Printf("Node was not found: %s", vt.votedFor)
	}

	node.State.vote = vt
}

func (r *Raft) releaseVotes() error {
	for _, node := range r.Nodes() {
		node.State = nil
	}

	return nil
}

// calculateVotes return true if majority --> Leader
// or returns false if not majority --> Candidate
func (r *Raft) calculateVotes() bool {
	online := 0
	votes := 0
	for _, node := range r.Nodes() {
		if node.State.vote != nil {
			online++
			if node.State.vote.voteGranted {
				votes++
			}
		}
	}

	if (votes + 1) >= (online+1)/2 {
		return true
	}

	return false
}

func (r *Raft) resetTicker() {
	r.electionTimeoutTickerLock.Lock()
	r.electionTimeoutTicker = getElectionTicker(r.config.ElectionTime)
	r.electionTimeoutTickerLock.Unlock()
}

func (r *Raft) stopTicker() {
	r.electionTimeoutTickerLock.Lock()
	r.electionTimeoutTicker.Stop()
	r.electionTimeoutTicker = nil
	r.electionTimeoutTickerLock.Unlock()
}

func (r *Raft) resetVotingTicker() {
	r.votingTimeoutTickerLock.Lock()
	r.votingTimeoutTicker = getElectionTicker(r.config.ElectionTime)
	r.votingTimeoutTickerLock.Unlock()
}

func (r *Raft) stopVotingTicker() {
	r.votingTimeoutTickerLock.Lock()
	r.votingTimeoutTicker.Stop()
	r.votingTimeoutTicker = nil
	r.votingTimeoutTickerLock.Unlock()
}
