package raft

import "os"

// RndTime ... Default 200
const (
	RndTime = 200
)

// Config will keep the configurations for n object
type Config struct {
	Name         string
	BindAddr     string
	BindTCPPort  int
	BindUDPPort  int
	ElectionTime int
	VotingTime   int
}

// DefaultConf returns the default configurations
func DefaultConf() *Config {
	hostname, _ := os.Hostname()

	config := &Config{
		Name:         hostname,
		BindAddr:     "127.0.0.1",
		BindTCPPort:  1111,
		BindUDPPort:  1111,
		ElectionTime: 500, // default 500
		VotingTime:   50,  // default 50
	}

	return config
}
