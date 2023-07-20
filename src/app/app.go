package app

import (
	"errors"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"stock-bot/src/archiver"
	"strings"
	"sync"
	"syscall"
)

// App builds the app with specified configuration.
func App() *cli.App {
	return &cli.App{
		Name:  "stock-archiver",
		Usage: "",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "symbols",
				Aliases:  []string{"s"},
				Usage:    "stock ticker symbols (can specify multiple)",
				Required: true,
			},
			&cli.IntFlag{
				Name:    "interval",
				Aliases: []string{"i"},
				Value:   1,
				Usage:   "check the symbols every interval seconds",
			},
		},
		Action: func(c *cli.Context) error {
			// Setup configuration
			cfg := Config{
				Symbols:  c.StringSlice("symbols"),
				Interval: c.Int("interval"),
			}

			// Setup stock-archiver
			{
				log.Print("STARTING: stock-archiver service")
				deduped := dedupeTickerSymbols(cfg.Symbols)
				if len(deduped) < 1 {
					return errors.New("error: must provide at least one ticker symbol")
				}
				if len(deduped) < len(cfg.Symbols) {
					log.Println("Removed duplicate ticker symbols.")
				}
				log.Printf("Starting up workers for %s", strings.Join(deduped, ", "))
				log.Printf("Running every %v seconds...", cfg.Interval)

				quitChan := make(chan struct{})
				wg := sync.WaitGroup{}
				for i, s := range deduped {
					wg.Add(1)
					wCfg := archiver.WorkerCfg{
						ID:         i,
						Symbol:     s,
						Interval:   cfg.Interval,
						SignalChan: quitChan,
						Group:      &wg,
					}
					go func(c archiver.WorkerCfg) {
						archiver.Worker(c)
					}(wCfg)
				}

				// Gracefully handle termination
				{
					interruptChan := make(chan os.Signal, 1)
					signal.Notify(interruptChan, os.Interrupt, syscall.SIGTERM)
					<-interruptChan
					log.Println("Stopping all workers, shutting down.")
					close(quitChan)
				}
				wg.Wait()
				log.Println("Stopped all workers.")
				return nil
			}
		},
	}
}

func dedupeTickerSymbols(symbols []string) []string {
	var deduped []string
	seen := make(map[string]bool)
	for _, s := range symbols {
		if _, in := seen[s]; !in {
			seen[s] = true
			deduped = append(deduped, s)
		}
	}
	return deduped
}
