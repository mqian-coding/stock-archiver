package archiver

import (
	"log"
	"sync"
	"time"
)

type WorkerCfg struct {
	ID         int
	Symbol     string
	Interval   int
	SignalChan chan struct{}
	Group      *sync.WaitGroup
}

func Worker(cfg WorkerCfg) {
	log.Printf("Starting worker: %v", cfg.ID)
	for {
		select {
		case <-cfg.SignalChan:
			log.Printf("Stopping Worker: %v", cfg.ID)
			cfg.Group.Done()
			return
		default:
			log.Printf("getting data for: %s", cfg.Symbol)
			time.Sleep(time.Duration(cfg.Interval) * time.Second)
		}
	}
}
