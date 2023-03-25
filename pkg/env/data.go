package env

import (
	"github.com/godoji/algocore/pkg/algo"
	"github.com/northberg/candlestick"
)

type MarketSupplier interface {
	Algorithm(name string, params ...float64) AlgorithmSupplier
	Interval(interval int64) IntervalSupplier
	Price() float64
	Time() int64
}

type IntervalSupplier interface {
	Candle() *candlestick.Candle
	FromLast(offset int) *candlestick.Candle
	ToTimeStamp(index int64) int64
	ToIndex(timeStamp int64) int64
	Indicator(name string, params ...int) IndicatorSupplier
}

type IndicatorSupplier interface {
	Value() float64
	Exists() bool
	Series(key string) float64
}

type AlgorithmSupplier interface {
	PastEvents() []*algo.Event
	CurrentEvents() []*algo.Event
	HasEvents() bool
}
