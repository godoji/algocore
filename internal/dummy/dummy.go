package dummy

import (
	"github.com/godoji/algocore/pkg/algo"
	"github.com/godoji/algocore/pkg/simulated"
	candles "github.com/northberg/candlestick"
)

var Params = make([]string, 0)

type CrossState = int

const (
	_ CrossState = iota
	StateUpTrend
	StateDownTrend
)

type LocalStore struct {
	State CrossState
}

// Evaluate Example algorithm for determining death-crosses on large time frames
func Evaluate(chart simulated.MarketSupplier, res *algo.ResultHandler, mem *simulated.Memory, param simulated.Parameters) {

	// A way of loading memory from disk
	var store *LocalStore
	if tmp := mem.Read(); tmp == nil {
		store = new(LocalStore)
	} else {
		store = tmp.(*LocalStore)
	}
	defer mem.Store(store)

	// Retrieve some indicators
	ema10 := chart.Interval(candles.Interval1d).Indicator("ema", 10).Value()
	ema50 := chart.Interval(candles.Interval1d).Indicator("ema", 50).Value()

	// Modify some states
	var nextState CrossState
	if ema50 > ema10 {
		nextState = StateDownTrend
	} else {
		nextState = StateUpTrend
	}

	// Create some events
	switch store.State {
	case StateUpTrend:
		if nextState == StateDownTrend {
			res.NewEvent("downtrend").SetColor("red").SetIcon("down")
		}
	case StateDownTrend:
		if nextState == StateUpTrend {
			res.NewEvent("uptrend").SetColor("green").SetIcon("up")
		}
	}

	// Store next state as current state
	store.State = nextState
}
