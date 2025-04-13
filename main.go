package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
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

	zone, offset := time.Now().Zone()
	log.Printf("[DEBUG] local timezone is %s (%+.02fh)", zone, float32(offset)/60/60)

	snmpclient, err := agentx.Dial("unix", CLI.SnmpMasterSock)
	if err != nil {
		log.Fatalf("Error connecting to SNMP master: %v", err)
	}
	snmpclient.Timeout = 1 * time.Minute
	snmpclient.ReconnectInterval = 1 * time.Second

	handler, err := NewBirdBGPHandler(CLI.BirdSock)
	if err != nil {
		log.Fatalf("Error initializing BGP handler: %v", err)
	}

	if err := handler.Register(CLI.SnmpPriority, snmpclient); err != nil {
		log.Fatalf("Error registering SNMP handler: %v", err)
	}

	log.Printf("[INFO] agentx started, waiting for requests")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(CLI.BirdRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := handler.Refresh(); err != nil {
				log.Printf("[ERROR] Failed to refresh BGP data: %v", err)
			}
		case sig := <-sigChan:
			log.Printf("[INFO] Received signal %v, shutting down", sig)
			return
		}
	}
}
