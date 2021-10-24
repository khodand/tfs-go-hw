package domain

import (
	"encoding/csv"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"sync"
)

type CandleCreator struct {
	writer1m io.Writer
	writer2m io.Writer
	writer10m io.Writer
}

func NewCandleCreator(w1m, w2m, w10m io.Writer) *CandleCreator {
	return &CandleCreator{
		writer1m: w1m,
		writer2m: w2m,
		writer10m: w10m,
	}
}

func (creator *CandleCreator) Process(wg *sync.WaitGroup, prices <-chan Price) {
	go func() {
		defer wg.Done()
		for range process(CandlePeriod10m, csv.NewWriter(creator.writer10m),
				process(CandlePeriod2m, csv.NewWriter(creator.writer2m),
				process(CandlePeriod1m,  csv.NewWriter(creator.writer1m), prices))) {}
	}()
}

func closeCandle(candle Candle, writer *csv.Writer){
	err := writer.Write([]string{string(candle.Ticker), candle.TS.String(),
		fmt.Sprintf("%f", candle.Open),
		fmt.Sprintf("%f", candle.High),
		fmt.Sprintf("%f", candle.Low),
		fmt.Sprintf("%f", candle.Close)})
	catchFatalError(err)
	writer.Flush()
}

func process(period CandlePeriod, w *csv.Writer, prices <-chan Price) <-chan Price {
	out := make(chan Price)
	activeCandles := make(map[TickerName]Candle)
	go func() {
		defer func() {
			for _, candle := range activeCandles {
				fmt.Println("defer ", period)
				closeCandle(candle, w)
			}
			close(out)
		}()

		for price := range prices {
			var candle Candle
			if c, ok := activeCandles[price.Ticker]; ok {
				candle = c
			} else {
				candle = NewCandle(price, period)
			}

			ts, err := PeriodTS(period, price.TS)
			catchFatalError(err)

			if ts.After(candle.TS) {
				closeCandle(candle, w)
				candle = NewCandle(price, period)
			} else {
				candle.UpdateCandle(price)
			}

			activeCandles[price.Ticker] = candle
			out <- price
			}
	}()

	return out
}

func catchFatalError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
