package domain

import (
	"context"
	"encoding/csv"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"sync"
)

type CandleCreator struct {
	ActiveCandles      map[TickerName]map[CandlePeriod]Candle
	ActiveCandlesMutex sync.Mutex

	writer1m io.Writer
	writer2m io.Writer
	writer10m io.Writer
}

func NewCandleCreator(w1m, w2m, w10m io.Writer) *CandleCreator {
	return &CandleCreator{
		ActiveCandles: make(map[TickerName]map[CandlePeriod]Candle),
		writer1m: w1m,
		writer2m: w2m,
		writer10m: w10m,
	}
}

func (creator *CandleCreator) Process(wg *sync.WaitGroup, ctx context.Context, prices <-chan Price) {
	go func() {
		defer wg.Done()
		lastOut := creator.process(CandlePeriod10m, creator.process(CandlePeriod2m, creator.process(CandlePeriod1m, prices)))
		for {
			select {
			case <-ctx.Done():
				return
			case <-lastOut:
				continue
			}
		}
	}()
}

func closeCandle(candle Candle, writer *csv.Writer){
	err := writer.Write([]string{string(candle.Ticker), candle.TS.String(),
		fmt.Sprintf("%f", candle.Open),
		fmt.Sprintf("%f", candle.High),
		fmt.Sprintf("%f", candle.Low),
		fmt.Sprintf("%f", candle.Close)})

	if err != nil {
		log.Fatalln("error writing csv:", err)
	}

	writer.Flush()
}

func (creator *CandleCreator) getCandle(price Price, period CandlePeriod) Candle{
	creator.ActiveCandlesMutex.Lock()
	defer creator.ActiveCandlesMutex.Unlock()
	if _, ok := creator.ActiveCandles[price.Ticker]; !ok {
		creator.ActiveCandles[price.Ticker] = make(map[CandlePeriod]Candle)
	}
	if candle, ok := creator.ActiveCandles[price.Ticker][period]; !ok {
		return NewCandle(price, period)
	} else {
		return candle
	}
}

func (creator *CandleCreator) setCandle(candle Candle, price Price, period CandlePeriod) {
	creator.ActiveCandlesMutex.Lock()
	creator.ActiveCandles[price.Ticker][period] = candle
	creator.ActiveCandlesMutex.Unlock()
}

func (creator *CandleCreator) process(period CandlePeriod, prices <-chan Price) <-chan Price {
	out := make(chan Price)
	var w *csv.Writer
	switch period {
	case CandlePeriod1m:
		w = csv.NewWriter(creator.writer1m)
	case CandlePeriod2m:
		w = csv.NewWriter(creator.writer2m)
	case CandlePeriod10m:
		w = csv.NewWriter(creator.writer10m)
	}

	go func() {
		defer func() {
			for _, periodMap := range creator.ActiveCandles {
				fmt.Println("defer ", period)
				closeCandle(periodMap[period], w)
			}
			close(out)
		}()


		for price := range prices {
			candle := creator.getCandle(price, period)

			ts, err := PeriodTS(period, price.TS)
			catchFatalError(err)

			if ts.After(candle.TS) {
				closeCandle(candle, w)
				candle = NewCandle(price, period)
			} else {
				candle.UpdateCandle(price)
			}

			creator.setCandle(candle, price, period)

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
