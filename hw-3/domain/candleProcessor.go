package domain

import (
	"context"
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
				creator.closeAllCandles()
				return
			case <-lastOut:
				continue
			}
		}
	}()
}

func (creator *CandleCreator) closeAllCandles() {
	for _, periodMap := range creator.ActiveCandles {
		for _, candle := range periodMap {
			fmt.Println(candle)
			creator.closeCandle(candle)
		}
	}
}

func (creator *CandleCreator) closeCandle(candle Candle){
	switch candle.Period {
	case CandlePeriod1m:
		candle.CloseCandle(creator.writer1m)
	case CandlePeriod2m:
		candle.CloseCandle(creator.writer2m)
	case CandlePeriod10m:
		candle.CloseCandle(creator.writer10m)
	}
}

func (creator *CandleCreator) getCandle(price Price, period CandlePeriod) Candle{
	if _, ok := creator.ActiveCandles[price.Ticker]; !ok {
		creator.ActiveCandles[price.Ticker] = make(map[CandlePeriod]Candle)
	}
	if candle, ok := creator.ActiveCandles[price.Ticker][period]; !ok {
		return NewCandle(price, period)
	} else {
		return candle
	}
}

func (creator *CandleCreator) process(period CandlePeriod, prices <-chan Price) <-chan Price {
	out := make(chan Price)
	go func() {
		defer close(out)
		for price := range prices {
			creator.ActiveCandlesMutex.Lock()
			candle := creator.getCandle(price, period)
			creator.ActiveCandlesMutex.Unlock()

			ts, err := PeriodTS(period, price.TS)
			catchFatalError(err)

			if ts.After(candle.TS) {
				creator.closeCandle(candle)
				candle = NewCandle(price, period)
			} else {
				candle.UpdateCandle(price)
			}

			creator.ActiveCandlesMutex.Lock()
			creator.ActiveCandles[price.Ticker][period] = candle
			creator.ActiveCandlesMutex.Unlock()

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
