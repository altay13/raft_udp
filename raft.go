package raft

import (
	"fmt"
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

	nodes   []*Node
	nodeMap map[string]*Node

	heartBeatLock sync.Mutex
	heartBeat     *time.Ticker
	stopHeartBeat chan bool

	electionTimeoutTickerLock sync.Mutex
	electionTimeoutTicker     *time.Ticker

	beartBeatChan sync.Mutex
	hbChan        chan *leaderHeartBeat // Altay check it

	votes chan int

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
		hbChan:       make(chan *leaderHeartBeat, 1),
		votes:        make(chan int),
		stateChanged: make(chan bool),
	}

	go r.ListenTCP()
	go r.ListenUDP()

	return r, nil
}

// Init func
func Init(conf *Config) (*Raft, error) {
	r, err := newRaft(conf)
	if err != nil {
		return nil, err
	}

	state, err := createInitState(conf)
	if err != nil {
		return nil, err
	}

	r.self = state

	return r, nil
}

func createInitState(conf *Config) (*State, error) {

	// name := fmt.Sprintf("%s:%d&%d", conf.BindAddr, conf.BindTCPPort, conf.BindUDPPort)

	state := &State{
		state:    Leader,
		leaderID: conf.Name,
		term:     0,
		vote:     0,
		votedFor: "",
	}

	return state, nil
}
