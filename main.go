package main

import (
	"log"
	"time"

	"github.com/alecthomas/kong"
	"github.com/posteo/go-agentx"
)

var CLI struct {
	BirdSock            string        `short:"s" help:"bird socket path" default:"/run/bird/bird.ctl"`
	BirdRefreshInterval time.Duration `short:"r" help:"bird data refresh interval" default:"3s"`
	SnmpMasterSock      string        `short:"x" help:"snmpd agentx master socket path" default:"/var/agentx/master"`
	SnmpPriority        byte          `short:"p" help:"snmpd registration priority" default:"127"`
}

func main() {
	kong.Parse(&CLI)

	snmpclient, err := agentx.Dial("unix", CLI.SnmpMasterSock)
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
	snmpclient.Timeout = 1 * time.Minute
	snmpclient.ReconnectInterval = 1 * time.Second

	handler, err := NewBirdBGPHandler(CLI.BirdSock)
	if err != nil {
		log.Fatalf("Error: %s", err)
		return
	}

	if err := handler.Register(CLI.SnmpPriority, snmpclient); err != nil {
		log.Fatalf("Error: %s", err)
		return
	}
	log.Printf("[INFO] agentx started, waiting for requests")
	ticker := time.NewTicker(CLI.BirdRefreshInterval)
	for range ticker.C {
		handler.Refresh()
	}
}
