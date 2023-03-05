package env

import "github.com/northberg/candlestick"

type MarketSupplier interface {
	Interval(interval int64) IntervalSupplier
	Price() float64
	Time() int64
}

type IntervalSupplier interface {
	Candle() *candlestick.Candle
	FromLast(offset int) *candlestick.Candle
	Indicator(name string, params ...int) IndicatorSupplier
}

type IndicatorSupplier interface {
	Value() float64
	Exists() bool
	Series(key string) float64
}
