package raft

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// Raft is the main object of the node itself
type Raft struct {
	config *Config

	tcpListener *net.TCPListener
	udpListener *net.UDPConn

	stateLock sync.Mutex
	self      *State

	stateCh chan int

	nodes   []*Node
	nodeMap map[string]*Node

	heartBeatLock sync.Mutex
	heartBeat     *time.Ticker
	stopHeartBeat chan bool

	electionTimeoutTickerLock sync.Mutex
	electionTimeoutTicker     *time.Ticker

	votingTimeoutTickerLock sync.Mutex
	votingTimeoutTicker     *time.Ticker

	beartBeatChan sync.Mutex
	hbChan        chan *heartBeat // Altay check it

	votes chan *Vote

	stateChanged chan bool
}

func newRaft(conf *Config) (*Raft, error) {
	tcplnAddr := fmt.Sprintf("%s:%d", conf.BindAddr, conf.BindTCPPort)
	tcpln, err := net.Listen("tcp", tcplnAddr)
	if err != nil {
		return nil, fmt.Errorf("Failed to start TCP listener. Err: %s", err)
	}

	udplnAddr := fmt.Sprintf("%s:%d", conf.BindAddr, conf.BindUDPPort)
	udpln, err := net.ListenPacket("udp", udplnAddr)
	if err != nil {
		return nil, fmt.Errorf("Failed to start UDP listener. Err: %s", err)
	}

	r := &Raft{
		config:       conf,
		tcpListener:  tcpln.(*net.TCPListener),
		udpListener:  udpln.(*net.UDPConn),
		nodeMap:      make(map[string]*Node),
		hbChan:       make(chan *heartBeat, 1),
		votes:        make(chan *Vote, 100),
		stateChanged: make(chan bool),
		stateCh:      make(chan int),
	}

	go r.ListenTCP()
	go r.ListenUDP()

	go r.handleState()

	return r, nil
}

// Init func
func Init(conf *Config) (*Raft, error) {
	r, err := newRaft(conf)
	if err != nil {
		return nil, err
	}

	st, err := createInitState(conf)
	if err != nil {
		return nil, err
	}

	r.self = st

	r.stateCh <- r.self.state

	return r, nil
}

func createInitState(conf *Config) (*State, error) {

	// name := fmt.Sprintf("%s:%d&%d", conf.BindAddr, conf.BindTCPPort, conf.BindUDPPort)

	state := &State{
		state:    Follower,
		leaderID: "",
		term:     0,
		vote: &Vote{
			votedFor:   "",
			voteStatus: UnknownVote,
		},
	}

	return state, nil
}

// Join func
func Join(conf *Config, joinAddr string, port int) (*Raft, error) {
	r, err := newRaft(conf)
	if err != nil {
		return nil, err
	}

	st, err := createInitState(conf)
	if err != nil {
		return nil, err
	}

	st.state = Unknown
	r.self = st

	addr := getUDPAddr(joinAddr, port)

	go r.joinToClusterByAddr(addr)

	return r, nil
}

// Nodes ...
func (r *Raft) Nodes() []*Node {
	return r.nodes
}

// Shutdown closes the listener and shuts down the Raft
func (r *Raft) Shutdown() {
	r.tcpListener.Close()
	r.udpListener.Close()

	log.Println("Shutting down the Raft...")
}

// GetState temporary function
func (r *Raft) GetState() {

	fmt.Println("Machine name: ", r.config.Name)

	fmt.Printf("Nodes: \n")
	for num, node := range r.Nodes() {
		fmt.Printf("%d : %s\n", num, node.Name)
	}

	switch r.self.state {
	case Follower:
		fmt.Println("I am Follower")
	case Leader:
		fmt.Println("I am Leader")
	case Candidate:
		fmt.Println("I am Candidate")
	}
}
