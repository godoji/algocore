package main

import (
	"github.com/godoji/algocore/internal/dummy"
	"github.com/godoji/algocore/pkg/ritmic"
	"testing"
)

func TestBasic(t *testing.T) {
	ritmic.RunTestShort(dummy.EvaluateLastCandle, [][]float64{{}}, dummy.ParamsLastCandle)
	ritmic.RunTestShort(dummy.EvaluateAnyCandles, [][]float64{{50}}, dummy.ParamsAnyCandles)
}

func TestLinked(t *testing.T) {
	ritmic.RunTestShort(dummy.EvaluateRecursive, [][]float64{{}}, dummy.ParamsRecursive)
}
