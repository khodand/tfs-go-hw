package domain

import (
	"encoding/csv"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"sync"
)

type CandleCreator struct {
	writer1m  io.Writer
	writer2m  io.Writer
	writer10m io.Writer
}

func NewCandleCreator(w1m, w2m, w10m io.Writer) *CandleCreator {
	return &CandleCreator{
		writer1m:  w1m,
		writer2m:  w2m,
		writer10m: w10m,
	}
}

func (creator *CandleCreator) Process(wg *sync.WaitGroup, prices <-chan Price) {
	go func() {
		defer wg.Done()
		for range save(csv.NewWriter(creator.writer10m), process(CandlePeriod10m,
			save(csv.NewWriter(creator.writer2m), process(CandlePeriod2m,
				save(csv.NewWriter(creator.writer1m), process(CandlePeriod1m, prices)))))) {
		}
	}()
}

func closeCandle(candle Candle, writer *csv.Writer) {
	err := writer.Write([]string{string(candle.Ticker), candle.TS.String(),
		fmt.Sprintf("%f", candle.Open),
		fmt.Sprintf("%f", candle.High),
		fmt.Sprintf("%f", candle.Low),
		fmt.Sprintf("%f", candle.Close)})
	catchFatalError(err)
	writer.Flush()
}

func save(w *csv.Writer, candles <-chan Candle) <-chan Price {
	out := make(chan Price)
	go func() {
		defer func() {
			close(out)
		}()

		for candle := range candles {
			if candle.Closed {
				closeCandle(candle, w)
			}
			out <- Price{
				Ticker: candle.Ticker,
				Value:  candle.Close,
				TS:     candle.TS,
			}
		}
	}()
	return out
}

func process(period CandlePeriod, prices <-chan Price) <-chan Candle {
	out := make(chan Candle)
	activeCandles := make(map[TickerName]Candle)
	go func() {
		defer func() {
			for _, candle := range activeCandles {
				candle.Closed = true
				out <- candle
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

			outCandle := candle
			if ts.After(candle.TS) {
				outCandle.Closed = true
				candle = NewCandle(price, period)
			} else {
				candle.UpdateCandle(price)
			}

			activeCandles[price.Ticker] = candle
			out <- outCandle
		}
	}()

	return out
}

func catchFatalError(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
