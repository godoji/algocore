package simulation

import (
	threading "github.com/aelbrecht/go-threader"
	"github.com/godoji/algocore/pkg/algo"
	"github.com/godoji/algocore/pkg/env"
	"github.com/godoji/algocore/pkg/kiosk"
	"github.com/northberg/candlestick"
	"runtime"
	"strings"
	"sync"
	"time"
)

type StepFunction = func(chart env.MarketSupplier, term *algo.ResultHandler, mem *env.Memory, params env.Parameters)

var metricsLock = sync.Mutex{}

type ResultWithLock struct {
	Data *algo.ResultSet
	Lock sync.Mutex
}

type Evaluator struct {
	step       StepFunction
	symbols    []candlestick.AssetIdentifier
	resolution int64
	maxThreads int
	metrics    algo.Status
	results    *algo.ResultSet
}

func (s *Evaluator) SetMaxThreads(threads int) *Evaluator {
	s.maxThreads = threads
	return s
}

func (s *Evaluator) Metrics() *algo.Status {
	return &s.metrics
}

func (s *Evaluator) Results() *algo.ResultSet {
	return s.results
}

type EvalOptions struct {
	Step       StepFunction
	Resolution int64
	Symbols    []string
}

func NewEvaluator(opts EvalOptions) *Evaluator {
	assets := make([]candlestick.AssetIdentifier, 0)
	for _, symbol := range opts.Symbols {
		// TODO: move this to candlestick lib
		xs := strings.Split(symbol, ":")
		assets = append(assets, candlestick.NewAssetIdentifier(xs[0], xs[1], xs[2]))
	}
	return &Evaluator{
		step:       opts.Step,
		symbols:    assets,
		resolution: opts.Resolution,
		maxThreads: runtime.NumCPU(),
		metrics:    algo.Status{},
		results:    nil,
	}
}

func (s *Evaluator) Run(scenarios [][]float64, keys []string) *Evaluator {

	// start timer
	s.metrics.StartTime = time.Now().UTC().UnixMilli()
	s.metrics.Running = true

	// prepare results
	results := &ResultWithLock{
		Data: &algo.ResultSet{
			Symbols: make(map[string]*algo.SymbolResultSet),
		},
	}

	// create tasks per segment and per symbol
	tasks := make([]*Task, 0)
	for _, symbol := range s.symbols {
		tasks = append(tasks, &Task{
			symbol: symbol,
		})
	}

	if s.maxThreads > 1 {
		threads := threading.NewThreader(s.maxThreads)
		for i := range tasks {
			task := tasks[i]
			threads.Run(func() {
				task.Simulate(s, scenarios, keys, results)
			})
		}
		threads.Wait()
	} else {
		for i := range tasks {
			task := tasks[i]
			task.Simulate(s, scenarios, keys, results)
		}
	}

	// save elapsed time
	s.metrics.Elapsed = time.Now().UTC().UnixMilli() - s.metrics.StartTime
	s.metrics.Running = false
	s.results = results.Data
	s.metrics.Finished = true

	return s
}

type Task struct {
	symbol candlestick.AssetIdentifier
}

func (s *Task) Simulate(sim *Evaluator, scenarios [][]float64, keys []string, results *ResultWithLock) {

	// provider for all scenarios
	provider := kiosk.NewProvider(s.symbol, sim.resolution)

	// parameters
	parameters := make([]env.Parameters, len(scenarios))
	for i := range parameters {
		parameters[i] = env.NewParameters(scenarios[i], keys)
	}

	// create memory for each scenario
	memories := make([]*env.Memory, len(scenarios))
	for i := range memories {
		memories[i] = env.NewMemory()
	}

	// create place to store all results
	resultSet := &algo.SymbolResultSet{
		Scenarios: make([]*algo.ScenarioSet, len(scenarios)),
	}
	results.Lock.Lock()
	results.Data.Symbols[s.symbol.ToString()] = resultSet
	results.Lock.Unlock()

	// store parameters in result set
	for i := range resultSet.Scenarios {
		resultSet.Scenarios[i] = &algo.ScenarioSet{
			Events:     make([]*algo.Event, 0),
			Parameters: scenarios[i],
		}
	}

	info := provider.Info()

	// iterate block per block, taking advantage of cached requests
	// TODO: move this to candlestick lib
	algoSupplier := kiosk.NewAlgorithmStore(s.symbol, sim.resolution)
	blockTimeSize := provider.Resolution() * candlestick.CandleSetSize
	startBlock := info.OnBoardDate / blockTimeSize
	currentBlock := time.Now().UTC().Unix() / blockTimeSize
	for block := startBlock; block <= currentBlock; block++ {

		// create data store for current block
		prev := provider.NewDataStore(block - 1)
		curr := provider.NewDataStore(block)

		// iterate 5000 minute candles
		for i := 0; i < 5000; i++ {

			// check if market is open
			candle := &curr.CandleSet(sim.resolution).Candles[i]
			if candle.Missing {
				continue
			}

			// create data supplier for current time instance
			ds := kiosk.NewSupplier(prev, curr, i, algoSupplier)

			// iterate scenarios
			for j := range scenarios {

				// retrieve memory
				mem := memories[j]

				// create handler for results
				res := algo.NewResultHandler(resultSet.Scenarios[j], ds.Time(), ds.Price())

				// evaluate trading script
				sim.step(&ds, res, mem, parameters[j])
			}
		}

	}

}
