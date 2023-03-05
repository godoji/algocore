package main

import (
	"github.com/godoji/algocore/internal/dummy"
	"github.com/godoji/algocore/pkg/ritmic"
)

// Run a server to use this algorithm in headless mode
func main() {
	ritmic.Serve(dummy.EvaluateLastCandle, dummy.ParamsLastCandle)
}
