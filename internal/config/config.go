package config

import (
	"os"
	"path/filepath"
)

const (
	DefaultGossipPort = 7946
	DefaultDaemonPort = 8080
	DefaultDNSPort    = 5353
	DefaultDataDir    = ".ghostmesh"
)

type Config struct {
	ClusterToken string
	GossipPort   int
	DaemonPort   int
	DNSPort      int
	DataDir      string
	NodeName     string
}

func Default() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		GossipPort: DefaultGossipPort,
		DaemonPort: DefaultDaemonPort,
		DNSPort:    DefaultDNSPort,
		DataDir:    filepath.Join(home, DefaultDataDir),
		NodeName:   hostname(),
	}
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}
