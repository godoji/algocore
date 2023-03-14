package kiosk

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"github.com/godoji/algocore/pkg/algo"
	"github.com/northberg/candlestick"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var kioUrl, incaUrl, algoUrl string

var (
	cache       *ristretto.Cache
	cacheLive   = false
	cacheConfig = ristretto.Config{
		NumCounters: 512,
		MaxCost:     1 << 29, // 512MB
		BufferItems: 64,
	}
	marketInfoCache     *candlestick.ExchangeList = nil
	marketInfoCacheLock                           = sync.Mutex{}
)

// setup caches
func init() {

	if len(os.Args) < 2 {
		return
	}
	mode := os.Args[1]
	if mode == "live" {
		cacheLive = true
	}

	var err error
	if cache, err = ristretto.NewCache(&cacheConfig); err != nil {
		log.Fatal(err)
	}
}

func init() {
	if v := os.Getenv("KIO_URL"); v != "" {
		kioUrl = v
	} else {
		log.Fatalln("KIO_URL not set")
	}
	if v := os.Getenv("INCA_URL"); v != "" {
		incaUrl = v
	} else {
		log.Fatalln("INCA_URL not set")
	}
	if v := os.Getenv("ALGO_URL"); v != "" {
		algoUrl = v
	} else {
		log.Fatalln("ALGO_URL not set")
	}
}

func getRequestProgress(url string) int {
	inProgressLock.Lock()
	progress := inProgress[url]
	inProgressLock.Unlock()
	return progress
}

func setRequestProgress(url string, state int) {
	inProgressLock.Lock()
	inProgress[url] = state
	inProgressLock.Unlock()
}

func fetch(url string, decoder func([]byte) (interface{}, error)) (interface{}, error) {

	// lock fetch queue
	inProgressLock.Lock()
	_, exists := inProgress[url] // check if fetch is in progress
	if !exists {
		inProgress[url] = 1 // mark url as being fetched
	}
	inProgressLock.Unlock()

	// do not send same request, wait
	if exists {
		for getRequestProgress(url) == 1 {
			time.Sleep(5 * time.Millisecond)
		}
	}

	if getRequestProgress(url) == 2 {
		c, ok := cache.Get(url)
		if ok {
			return c, nil
		} else {
			setRequestProgress(url, 1)
		}
	}

	// setup request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(url)
		log.Printf("failed to initialize get request\n")
		log.Fatalln(err)
	}

	// set headers
	req.Header.Set("Accept", "application/octet-stream")

	// execute request and handle any connection or url based error
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(url)
		log.Printf("failed to fetch, network error\n")
		log.Fatalln(err)
	}

	// case when no candle data exists
	if resp.StatusCode == http.StatusNotFound {
		if cacheLive {
			cache.SetWithTTL(url, nil, 1, time.Second)
		} else {
			cache.Set(url, nil, 1)
		}
		return nil, nil
	}

	// check if response is useful
	if resp.StatusCode != http.StatusOK {
		log.Println(url)
		log.Fatalf("indicator request failed with code %d\n", resp.StatusCode)
	}

	if resp.Body == nil {
		log.Println(url)
		log.Fatalf("decoding data failed\n")
	}

	// read data
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(url)
		log.Fatalf("reading payload failed: %s\n", err.Error())
	}

	// decode data
	e, err := decoder(body)
	if err != nil {
		log.Println(url)
		log.Fatalf("decoding data failed: %s\n", err.Error())
	}

	// drain request
	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		log.Fatal(err)
	}

	if err = resp.Body.Close(); err != nil {
		log.Fatal(err)
	}

	// update cache
	if cacheLive {
		cache.SetWithTTL(url, e, 1<<19, time.Second)
	} else {
		cache.Set(url, e, 1<<19)
	}

	// update get status
	inProgressLock.Lock()
	inProgress[url] = 2
	inProgressLock.Unlock()

	return e, nil
}

var inProgress = make(map[string]int)
var inProgressLock = sync.Mutex{}

func GetCandles(block int64, interval int64, resolution int64, symbol string) (*candlestick.CandleSet, error) {

	// cache
	cacheParam := ""
	if cacheLive {
		cacheParam = "&cache=no-cache"
	}

	// fetch
	url := fmt.Sprintf("%s/market/t/%s?segment=%d&interval=%d&resolution=%d%s",
		kioUrl, symbol, block, interval, resolution, cacheParam)
	raw, err := fetch(url, func(b []byte) (interface{}, error) { return candlestick.DecodeCandleSet(b) })
	if err != nil {
		return nil, err
	}

	// cast
	var result *candlestick.CandleSet = nil
	if raw != nil {
		result = raw.(*candlestick.CandleSet)
	}
	return result, nil

}

func GetIndicator(block int64, name string, interval int64, resolution int64, symbol string, params []int) (*candlestick.Indicator, error) {

	// cache
	cacheParam := ""
	if cacheLive {
		cacheParam = "&cache=no-cache"
	}

	// fetch
	url := fmt.Sprintf("%s/indicators/t/%s?block=%d&interval=%d&resolution=%d&symbol=%s&params=%s%s",
		incaUrl, name, block, interval, resolution, symbol, concatParams(params), cacheParam)
	raw, err := fetch(url, func(b []byte) (interface{}, error) { return candlestick.DecodeIndicatorSet(b) })
	if err != nil {
		return nil, err
	}

	// cast
	var result *candlestick.Indicator = nil
	if raw != nil {
		result = raw.(*candlestick.Indicator)
	}
	return result, nil

}

func GetAlgorithm(name string, resolution int64, symbol string, params []float64) (*algo.ScenarioResultSet, error) {

	// cache
	cacheParam := ""
	if cacheLive {
		cacheParam = "&cache=no-cache"
	}

	// fetch
	url := fmt.Sprintf("%s/sync/algorithms/%s?resolution=%d&symbol=%s&params=%s%s",
		algoUrl, name, resolution, symbol, concatParamsFloat(params), cacheParam)

	// setup request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(url)
		log.Printf("failed to initialize get request\n")
		log.Fatalln(err)
	}

	// set headers
	req.Header.Set("Accept", "application/octet-stream")

	// execute request and handle any connection or url based error
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(url)
		log.Printf("failed to fetch, network error\n")
		log.Fatalln(err)
	}

	// case when no candle data exists
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	// check if response is useful
	if resp.StatusCode != http.StatusOK {
		log.Println(url)
		log.Fatalf("indicator request failed with code %d\n", resp.StatusCode)
	}

	if resp.Body == nil {
		log.Println(url)
		log.Fatalf("decoding data failed\n")
	}

	// read data
	result := new(algo.ScenarioResultSet)
	err = gob.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		log.Println(url)
		log.Fatalf("reading payload failed: %s\n", err.Error())
	}

	// drain request
	if _, err = io.Copy(io.Discard, resp.Body); err != nil {
		log.Fatal(err)
	}

	if err = resp.Body.Close(); err != nil {
		log.Fatal(err)
	}

	return result, nil
}

func GetExchangeInfo() (*candlestick.ExchangeList, error) {

	marketInfoCacheLock.Lock()
	defer marketInfoCacheLock.Unlock()

	if marketInfoCache == nil {

		// fetch
		resp, err := http.Get(fmt.Sprintf("%s/market/info", kioUrl))
		if err != nil {
			return nil, err
		}

		// decode
		result := new(candlestick.ExchangeList)
		err = json.NewDecoder(resp.Body).Decode(result)

		// drain and close body
		if resp.Body != nil {
			if _, err := io.Copy(io.Discard, resp.Body); err != nil {
				log.Print(err)
			}
			if err := resp.Body.Close(); err != nil {
				log.Print(err)
			}
		}

		if err != nil {
			return nil, err
		}

		marketInfoCache = result
	}

	return marketInfoCache, nil

}

func concatParams(params []int) string {
	res := ""
	for i := range params {
		if i != 0 {
			res += ","
		}
		res += strconv.Itoa(params[i])
	}
	return res
}

func concatParamsFloat(params []float64) string {
	res := ""
	for i := range params {
		if i != 0 {
			res += ","
		}
		res += strconv.FormatFloat(params[i], 'f', -1, 64)
	}
	return res
}
