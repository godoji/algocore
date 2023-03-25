package kiosk

import (
	"github.com/godoji/algocore/pkg/algo"
	"github.com/godoji/algocore/pkg/env"
	"github.com/northberg/candlestick"
	"log"
	"sync"
)

type IntervalSupplier struct {
	parent   *DataSupplier
	interval int64
}

func (s *DataSupplier) Interval(interval int64) env.IntervalSupplier {
	return IntervalSupplier{
		parent:   s,
		interval: interval,
	}
}

func (s *DataSupplier) Price() float64 {
	return s.Interval(s.curr.provider.resolution).Candle().Close
}

func (s *DataSupplier) Time() int64 {
	return s.Interval(s.curr.provider.resolution).Candle().Time
}

type DataSupplier struct {
	index      int
	curr       *DataStore
	prev       *DataStore
	algorithms *AlgorithmStore
}

type DataStore struct {
	provider      *Provider
	block         int64
	candles       map[int64]*candlestick.CandleSet
	indicatorLock sync.Mutex
	indicators    map[string]*IndicatorSubStore
}

type IndicatorSubStore struct {
	Data map[int]*ParamSubStore
	Lock sync.Mutex
}

type ParamSubStore struct {
	Data []*candlestick.Indicator
	Lock sync.Mutex
}

type AlgorithmSubStore struct {
	Data []*algo.ScenarioSet
	Lock sync.Mutex
}

func (s *DataStore) fetchIndicator(name string, interval int64, params []int) *candlestick.Indicator {
	var err error
	indicator, err := GetIndicator(s.block, name, interval, s.provider.resolution, s.provider.symbol.ToString(), params)
	if err != nil {
		log.Fatalln(err)
	}
	if indicator == nil {
		log.Fatalf("indicator \"%s\" does not exist\n", name)
	}
	return indicator
}

func (s *DataStore) Indicator(name string, interval int64, params []int) *candlestick.Indicator {

	// retrieve map of indicators
	s.indicatorLock.Lock()
	subStore, ok := s.indicators[name]
	if !ok {
		subStore = &IndicatorSubStore{Data: make(map[int]*ParamSubStore)}
		s.indicators[name] = subStore
	}
	s.indicatorLock.Unlock()

	// find bucket based on params
	var arr *ParamSubStore
	key := 0
	if len(params) > 0 {
		key = params[0]
	}

	subStore.Lock.Lock()
	arr, ok = subStore.Data[key]
	if !ok {
		arr = &ParamSubStore{Data: make([]*candlestick.Indicator, 0)}
		subStore.Data[key] = arr
	}
	subStore.Lock.Unlock()

	// look in bucket for indicator
	arr.Lock.Lock()
	for _, v := range arr.Data {
		if v.Meta.Name != name {
			panic("wrong indicator in sub-store")
		}
		if v.Meta.BaseInterval != interval {
			panic("wrong interval in sub-store")
		}
		if len(v.Meta.Parameters) != len(params) {
			continue
		}
		invalid := false
		for j, p := range v.Meta.Parameters {
			if p != params[j] {
				invalid = true
				break
			}
		}
		if !invalid {
			arr.Lock.Unlock()
			return v
		}
	}

	// fetch and add to bucket if not found
	indicator := s.fetchIndicator(name, interval, params)
	arr.Data = append(arr.Data, indicator)
	arr.Lock.Unlock()

	return indicator
}

func (s *AlgorithmStore) fetchAlgorithm(name string, params []float64) *algo.ScenarioSet {
	var err error
	result, err := GetAlgorithm(name, s.resolution, s.symbol.ToString(), params)
	if err != nil {
		log.Fatalln(err)
	}
	if result == nil {
		log.Fatalf("algorithm \"%s\" does not exist\n", name)
	}
	return result
}

func (s *AlgorithmStore) algorithm(name string, params []float64) *algo.ScenarioSet {

	// retrieve map of indicators
	s.algorithmLock.Lock()
	arr, ok := s.algorithms[name]
	if !ok {
		arr = &AlgorithmSubStore{Data: make([]*algo.ScenarioSet, 0)}
		s.algorithms[name] = arr
	}
	s.algorithmLock.Unlock()

	// look in bucket for indicator
	arr.Lock.Lock()
	for _, v := range arr.Data {
		if len(v.Parameters) != len(params) {
			continue
		}
		invalid := false
		for j, p := range v.Parameters {
			if p != params[j] {
				invalid = true
				break
			}
		}
		if !invalid {
			arr.Lock.Unlock()
			return v
		}
	}

	// fetch and add to bucket if not found
	result := s.fetchAlgorithm(name, params)
	arr.Data = append(arr.Data, result)
	arr.Lock.Unlock()

	if len(result.Parameters) != len(params) {
		log.Println("parameter mismatch: parameters must be passed explicitly")
		log.Fatalf("expected %d parameters but got %d instead", len(params), len(result.Parameters))
	}

	return result
}

func (s *DataStore) CandleSet(interval int64) *candlestick.CandleSet {
	candles, ok := s.candles[interval]
	if !ok {
		var err error
		candles, err = GetCandles(s.block, interval, s.provider.resolution, s.provider.symbol.ToString())
		if err != nil {
			log.Fatalln(err)
		}
		if candles == nil {
			log.Fatalf("failed to fetch %s candles for block %d (%d)\n", s.provider.symbol.ToString(), s.block, interval)
		}
		s.candles[interval] = candles
	}
	return candles
}

type Provider struct {
	symbol     candlestick.AssetIdentifier
	resolution int64
}

func NewProvider(symbol candlestick.AssetIdentifier, resolution int64) *Provider {
	return &Provider{
		symbol:     symbol,
		resolution: resolution,
	}
}

func (p *Provider) NewDataStore(block int64) *DataStore {
	return &DataStore{
		provider:   p,
		block:      block,
		candles:    make(map[int64]*candlestick.CandleSet),
		indicators: make(map[string]*IndicatorSubStore),
	}
}

func (p *Provider) Resolution() int64 {
	return p.resolution
}

func (p *Provider) Info() *candlestick.AssetInfo {
	exchangeInfo, err := GetExchangeInfo()
	if err != nil {
		log.Println("failed to retrieve exchange info")
		log.Fatalln(err)
	}
	for _, exchange := range exchangeInfo.Exchanges {
		if exchange.BrokerId != p.symbol.Broker {
			continue
		}
		info, ok := exchange.Symbol(p.symbol.ToString())
		if !ok {
			continue
		}
		return info
	}
	log.Fatalf("could not find broker for asset: %s\n", p.symbol.ToString())
	return nil
}

func NewSupplier(prev *DataStore, curr *DataStore, index int, alg *AlgorithmStore) DataSupplier {
	return DataSupplier{
		curr:       curr,
		prev:       prev,
		index:      index,
		algorithms: alg,
	}
}

func (s IntervalSupplier) Candle() *candlestick.Candle {
	return &s.parent.curr.CandleSet(s.interval).Candles[s.parent.index]
}

func (s IntervalSupplier) ToIndex(timeStamp int64) int64 {
	return -(s.parent.curr.CandleSet(s.interval).Index(timeStamp) - int64(s.parent.index))
}

func (s IntervalSupplier) ToTimeStamp(index int64) int64 {
	if index > 0 {
		log.Fatalln("cannot look into the future")
	}
	return s.parent.curr.CandleSet(s.interval).TimeStampAtIndex(-index + int64(s.parent.index))
}

func (s IntervalSupplier) Indicator(name string, params ...int) env.IndicatorSupplier {
	return IndicatorSupplier{
		name:      name,
		parent:    s.parent,
		indicator: s.parent.curr.Indicator(name, s.interval, params),
	}
}

func (s IntervalSupplier) FromLast(offset int) *candlestick.Candle {
	if offset < 0 {
		log.Fatalln("time offset cannot be negative")
	}
	if int64(offset) > candlestick.CandleSetSize {
		log.Fatalln("cannot retrieve candles more than 1 block back in time")
	}
	index := s.parent.index - offset
	ds := s.parent.curr
	if index < 0 {
		index += int(candlestick.CandleSetSize)
		ds = s.parent.prev
	}
	return &ds.CandleSet(s.interval).Candles[index]
}

type AlgorithmStore struct {
	algorithmLock sync.Mutex
	algorithms    map[string]*AlgorithmSubStore
	resolution    int64
	symbol        candlestick.AssetIdentifier
}

func NewAlgorithmStore(symbol candlestick.AssetIdentifier, resolution int64) *AlgorithmStore {
	return &AlgorithmStore{
		algorithms: map[string]*AlgorithmSubStore{},
		resolution: resolution,
		symbol:     symbol,
	}
}

func (s *DataSupplier) Algorithm(name string, params ...float64) env.AlgorithmSupplier {
	return AlgorithmSupplier{
		name:     name,
		parent:   s,
		scenario: s.algorithms.algorithm(name, params),
	}
}

type AlgorithmSupplier struct {
	name     string
	parent   *DataSupplier
	scenario *algo.ScenarioSet
}

func (s AlgorithmSupplier) HasEvents() bool {
	stepEnd := s.parent.Time()
	stepStart := stepEnd - s.parent.algorithms.resolution
	for _, event := range s.scenario.Events {
		if event.CreatedOn > stepStart && event.CreatedOn <= stepEnd {
			return true
		}
		if event.CreatedOn > stepEnd {
			return false
		}
	}
	return false
}

func (s AlgorithmSupplier) PastEvents() []*algo.Event {
	stepEnd := s.parent.Time()
	stepStart := stepEnd - s.parent.algorithms.resolution
	results := make([]*algo.Event, 0)
	for _, event := range s.scenario.Events {
		if event.CreatedOn <= stepStart {
			results = append(results, event)
		}
	}
	return results
}

func (s AlgorithmSupplier) CurrentEvents() []*algo.Event {
	stepEnd := s.parent.Time()
	stepStart := stepEnd - s.parent.algorithms.resolution
	results := make([]*algo.Event, 0)
	for _, event := range s.scenario.Events {
		if event.CreatedOn > stepStart && event.CreatedOn <= stepEnd {
			results = append(results, event)
		}
	}
	return results
}

type IndicatorSupplier struct {
	name      string
	parent    *DataSupplier
	indicator *candlestick.Indicator
}

func (s IndicatorSupplier) Exists() bool {
	for _, series := range s.indicator.Series {
		if series.Values[s.parent.index].Missing {
			return false
		}
	}
	return true
}

func (s IndicatorSupplier) Value() float64 {
	return s.Series(s.name)
}

func (s IndicatorSupplier) Series(key string) float64 {
	v, ok := s.indicator.Series[key]
	if !ok {
		log.Fatalf("indicator series \"%s\" does not exist in \"%s\"\n", key, s.name)
	}
	return v.Values[s.parent.index].Value
}
