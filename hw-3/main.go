package main

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hw-3/domain"
	"hw-3/generator"
	"os"
	"os/signal"
	"sync"
	"time"
)

var tickers = []domain.TickerName{"AAPL", "SBER", "NVDA", "TSLA"}

func main() {
	logger := log.New()
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	var shutdownCh = make(chan struct{})
	go func() {
		for {
			select {
			case <-shutdownCh:
				cancel()
				wg.Wait()
				return
			default:
			}
		}
	}()

	pg := generator.NewPricesGenerator(generator.Config{
		Factor:  10,
		Delay:   time.Millisecond * 500,
		Tickers: tickers,
	})

	file1m, _ := os.Create(fmt.Sprintf("candel_%s.csv", domain.CandlePeriod1m))
	defer file1m.Close()
	file2m, _ := os.Create(fmt.Sprintf("candel_%s.csv", domain.CandlePeriod2m))
	defer file2m.Close()
	file10m, _ := os.Create(fmt.Sprintf("candel_%s.csv", domain.CandlePeriod10m))
	defer file10m.Close()


	cp := domain.NewCandleCreator(file1m, file2m, file10m)

	os.Create("candel")
	logger.Info("start prices generator...")
	prices := pg.Prices(ctx)

	wg.Add(1)
	cp.Process(&wg, ctx, prices)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	close(shutdownCh)
	wg.Wait()
	cancel()
}
