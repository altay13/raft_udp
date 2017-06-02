package raft

import (
	"fmt"
	"log"
	"net"
	"time"
)

// First send ping to addr
// Wait for the response (have to define the timeout for that)
// if response then call RPC for getting the nodes and leader information
func (r *Raft) joinToClusterByAddr(addr net.Addr) error {

	joinTimeout := time.After(100 * time.Millisecond)

	for {
		select {
		case <-joinTimeout:
			if len(r.self.leaderID) > 0 {
				fmt.Println("Got the leader, I am leaving the loop")
				r.stateCh <- r.self.state
				return nil
			}

			p := &joinPing{
				Node: &Node{
					Name: r.config.Name,
					Addr: getUDPAddr(r.config.BindAddr, r.config.BindUDPPort),
				},
			}

			err := r.encodeAndSendMsg(addr, joinPingMsg, p)
			if err != nil {
				log.Printf("Failed to send the ping message. Err %s", err)
				return err
			}

			joinTimeout = time.After(1 * time.Second)
		}
	}

}

func (r *Raft) syncStateAfterJoin(dt *joinAckResp) error {
	err := r.addNode(dt.Node)

	if err != nil {
		log.Printf("Failed to add node to node list. Err %s", err)
		return err
	}

	for _, node := range dt.Nodes {
		err := r.addNode(node)
		if err != nil {
			log.Printf("Failed to add node to node list. Err %s", err)
			return err
		}
	}

	if r.self.state == Unknown {
		r.stateLock.Lock()
		r.self.leaderID = dt.LeaderID
		r.self.state = Follower
		r.stateLock.Unlock()
	}

	return nil

}

func (r *Raft) addNode(p *Node) error {
	if p.Name == r.config.Name {
		return nil
	}

	node := r.nodeMap[p.Name]

	if node != nil {
		return nil
	}

	node = &Node{
		Name:       p.Name,
		Addr:       p.Addr,
		VoteStatus: UnknownVote,
	}

	r.nodeMap[node.Name] = node

	r.nodes = append(r.nodes, node)

	return nil
}

func (r *Raft) invokeJoinAckResp(from net.Addr, p *joinPing) error {
	if r.self.state == Leader {
		ack := &joinAckResp{}
		node := &Node{
			Name: r.config.Name,
			Addr: getUDPAddr(r.config.BindAddr, r.config.BindUDPPort),
		}

		ack.LeaderID = r.self.leaderID
		ack.Nodes = r.Nodes()
		ack.Node = node

		err := r.encodeAndSendMsg(p.Node.Addr, joinAckRespMsg, &ack)
		if err != nil {
			log.Printf("Failed to send the ack to the ping. Err %s", err)
			return err
		}

		err = r.broadCastNewMember(p)
		if err != nil {
			log.Printf("Failed to broadcast about new node. Err %s", err)
		}

	} else {
		// forward the ping to server
		if len(r.self.leaderID) <= 0 {
			return nil
		}
		leaderAddr := r.nodeMap[r.self.leaderID].Addr
		if len(r.self.leaderID) > 0 {
			if leaderAddr.String() == from.String() {
				return nil
			}

			fmt.Println("Sent request to leader server: --> ", leaderAddr)
			err := r.encodeAndSendMsg(leaderAddr, joinPingMsg, &p)
			if err != nil {
				log.Printf("Failed to forward the ping to the leader. Err %s", err)
				return err
			}
		}
	}

	return nil
}

func (r *Raft) broadCastNewMember(p *joinPing) error {

	for _, node := range r.nodes {
		go r.encodeAndSendMsg(node.Addr, joinPingMsg, &p)
	}

	return nil
}
