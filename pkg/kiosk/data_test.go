package kiosk

import (
	"github.com/northberg/candlestick"
	"testing"
)

func TestGetAllCandles(t *testing.T) {
	collection, err := GetAllCandles(candlestick.Interval1d, candlestick.Interval1d, "UNICORN:US:KO")
	if err != nil {
		panic(err)
	}
	for _, set := range collection {
		if set == nil {
			panic("candle set is nil")
		}
	}
}
