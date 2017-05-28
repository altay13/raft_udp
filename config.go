package raft

import "os"

// Config will keep the configurations for n object
type Config struct {
	Name         string
	BindAddr     string
	BindTCPPort  int
	BindUDPPort  int
	ElectionTime int
}

// DefaultConf returns the default configurations
func DefaultConf() *Config {
	hostname, _ := os.Hostname()

	config := &Config{
		Name:         hostname,
		BindAddr:     "127.0.0.1",
		BindTCPPort:  1234,
		BindUDPPort:  1234,
		ElectionTime: 3,
	}

	return config
}
