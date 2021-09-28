package domain

import (
	"encoding/csv"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

type TickerName string

type Price struct {
	Ticker TickerName
	Value  float64
	TS     time.Time
}

var ErrUnknownPeriod = errors.New("unknown period")

type CandlePeriod string

const (
	CandlePeriod1m  CandlePeriod = "1m"
	CandlePeriod2m  CandlePeriod = "2m"
	CandlePeriod10m CandlePeriod = "10m"
)

func PeriodTS(period CandlePeriod, ts time.Time) (time.Time, error) {
	switch period {
	case CandlePeriod1m:
		return ts.Truncate(time.Minute), nil
	case CandlePeriod2m:
		return ts.Truncate(2 * time.Minute), nil
	case CandlePeriod10m:
		return ts.Truncate(10 * time.Minute), nil
	default:
		return time.Time{}, ErrUnknownPeriod
	}
}

type Candle struct {
	Ticker TickerName
	Period CandlePeriod // Интервал
	Open   float64      // Цена открытия
	High   float64      // Максимальная цена
	Low    float64      // Минимальная цена
	Close  float64      // Цена закрытие
	TS     time.Time    // Время начала интервала
}

func NewCandle(price Price, period CandlePeriod) Candle{
	if period != "" {
		log.Info("NewCandle", period, price)
	}
	if ts, err := PeriodTS(period, price.TS); err != nil {
		return Candle{}
	} else {
		return Candle{
			Ticker: price.Ticker,
			Period: period,
			Open:   price.Value,
			High:   price.Value,
			Low:    price.Value,
			Close:  price.Value,
			TS:     ts,
		}
	}
}

func (candle *Candle) UpdateCandle(price Price) {
	candle.Low = minFloat64(candle.Low, price.Value)
	candle.High = maxFloat64(candle.High, price.Value)
	candle.Close = price.Value
}

func (candle *Candle) CloseCandle(file io.Writer) {
	// We never write to the same file in parallel, so no locks
	w := csv.NewWriter(file)
	err := w.Write([]string{string(candle.Ticker), candle.TS.String(),
		fmt.Sprintf("%f", candle.Open),
		fmt.Sprintf("%f", candle.High),
		fmt.Sprintf("%f", candle.Low),
		fmt.Sprintf("%f", candle.Close)})

	if err != nil {
		log.Fatalln("error writing csv:", err)
	}
	w.Flush()
}

func minFloat64(a, b float64) float64{
	if a < b {
		return a
	} else {
		return b
	}
}

func maxFloat64(a, b float64) float64{
	if a > b {
		return a
	} else {
		return b
	}
}
