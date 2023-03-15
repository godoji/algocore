package dummy

import (
	"github.com/godoji/algocore/pkg/algo"
	"github.com/godoji/algocore/pkg/env"
	candles "github.com/northberg/candlestick"
)

var ParamsLastCandle = make([]string, 0)
var ParamsAnyCandles = []string{"historySize"}
var ParamsRecursive = make([]string, 0)

type CrossState = int

const (
	_ CrossState = iota
	StateUpTrend
	StateDownTrend
)

type LocalStoreLastCandle struct {
	State CrossState
}

func EvaluateLastCandle(chart env.MarketSupplier, res *algo.ResultHandler, mem *env.Memory, param env.Parameters) {

	// A way of loading memory from disk
	var store *LocalStoreLastCandle
	if tmp := mem.Read(); tmp == nil {
		store = new(LocalStoreLastCandle)
	} else {
		store = tmp.(*LocalStoreLastCandle)
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

type LocalStoreAnyCandles struct {
	State       CrossState
	Initialized bool
	History     *env.FiLoStack
}

func EvaluateAnyCandles(chart env.MarketSupplier, res *algo.ResultHandler, mem *env.Memory, param env.Parameters) {

	// A way of loading memory from disk
	var store *LocalStoreAnyCandles
	if tmp := mem.Read(); tmp == nil {
		store = new(LocalStoreAnyCandles)
	} else {
		store = tmp.(*LocalStoreAnyCandles)
	}
	defer mem.Store(store)

	// Initialize any memory, or append
	if !store.Initialized {
		histSize := param.GetInt("historySize")
		store.History = env.NewFiLoStack(histSize)
		for i := 0; i < histSize; i++ {
			store.History.Push(chart.Interval(candles.Interval1d).FromLast(i))
		}
		store.Initialized = true
	} else {
		store.History.Push(chart.Interval(candles.Interval1d).Candle())
	}

	// Calculate local maxima
	maxIndex := 0
	maxCandle := chart.Interval(candles.Interval1d).Candle()
	for i := 0; i < store.History.Size(); i++ {
		c := store.History.At(i).(*candles.Candle)
		if c.Close > maxCandle.Close {
			maxIndex = i
			maxCandle = c
		}
	}

	// Create some events
	if maxIndex == 0 {
		res.NewEvent("uptrend").SetColor("green").SetIcon("up")
	}
}

func EvaluateRecursive(chart env.MarketSupplier, res *algo.ResultHandler, mem *env.Memory, param env.Parameters) {
	highsAndLows := chart.Algorithm("highs-and-lows", 7)
	if !highsAndLows.HasEvents() {
		return
	}
}
