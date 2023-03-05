package main

import (
	"github.com/godoji/algocore/internal/dummy"
	"github.com/godoji/algocore/pkg/ritmic"
	"testing"
)

func TestBot(t *testing.T) {
	// check if bot actually works before deployment
	// this makes use of kio and inca, make sure env variables KIO_URL and INCA_URL are set
	ritmic.RunTestShort(dummy.EvaluateLastCandle, [][]float64{{}}, dummy.ParamsLastCandle)
	ritmic.RunTestShort(dummy.EvaluateAnyCandles, [][]float64{{50}}, dummy.ParamsAnyCandles)
}
