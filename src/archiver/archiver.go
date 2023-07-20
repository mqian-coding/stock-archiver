package archiver

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/piquette/finance-go/quote"
)

const CSV_STORAGE_PATH = "out/"

type WorkerCfg struct {
	ID         int
	Symbol     string
	Interval   int
	SignalChan chan struct{}
	Group      *sync.WaitGroup
}

func Worker(cfg WorkerCfg) {
	log.Printf("Starting worker: %v", cfg.ID)

	file, err := makeNewCSV(getCSVName(cfg.Symbol))
	if err != nil {
		log.Printf("error: failed to make the csv")
		panic(err)
	}
	w := csv.NewWriter(file)
	d := time.Now().Day()

	for {
		select {
		case <-cfg.SignalChan:
			log.Printf("Stopping Worker: %v", cfg.ID)
			cfg.Group.Done()
			return
		default:
			// Create new csv if midnight
			if shouldMakeNewCSV(d) {
				if err = file.Close(); err != nil {
					log.Printf("error: failed to close the csv")
					panic(err)
				}
				// Reset with freshly made csv
				file, err = makeNewCSV(getCSVName(cfg.Symbol))
				if err != nil {
					log.Printf("error: failed to make the csv")
					panic(err)
				}
				w = csv.NewWriter(file)
				if err = writeHeader(w); err != nil {
					log.Printf("error: failed to write the csv header")
					panic(err)
				}
				d = time.Now().Day()
			}

			if err = logQuotes(cfg.Symbol, w); err != nil {
				log.Print(err)
			}
			time.Sleep(time.Duration(cfg.Interval) * time.Second)
		}
	}
}

func getCSVName(symbol string) string {
	y, m, d := time.Now().Date()
	return fmt.Sprintf("%s_%v_%v_%v.csv", symbol, y, m, d)
}

func makeNewCSV(name string) (*os.File, error) {
	file, err := os.Create(CSV_STORAGE_PATH + name)
	if err != nil {
		return nil, err
	}
	return file, err
}

func logQuotes(symbol string, w *csv.Writer) error {
	q, err := quote.Get(symbol)
	if err != nil || q == nil {
		return err
	}
	row := []string{
		strconv.FormatFloat(q.Bid, 'E', -1, 64),
		strconv.FormatFloat(q.Ask, 'E', -1, 64),
		strconv.FormatFloat(q.RegularMarketPrice, 'E', -1, 64)}
	if err = writeRow(w, row); err != nil {
		return err
	}
	return nil
}

func writeHeader(w *csv.Writer) error {
	if err := w.Write([]string{"bid", "ask", "regular_market_price"}); err != nil {
		return err
	}
	w.Flush()
	return nil
}

func writeRow(w *csv.Writer, row []string) error {
	if err := w.Write(row); err != nil {
		return err
	}
	w.Flush()
	return nil
}

func shouldMakeNewCSV(day int) bool {
	return time.Now().Day() != day
}
