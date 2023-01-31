package ritmic

import (
	"encoding/json"
	"fmt"
	"github.com/godoji/algocore/internal/handler"
	"github.com/godoji/algocore/pkg/algo"
	"github.com/northberg/candlestick"
	"log"
)

func RunShortTestSet(step algo.StepFunction, scenarios [][]float64, paramKeys []string) *handler.Evaluator {
	fmt.Println("Running single-threaded test")
	bot := handler.NewEvaluator(handler.EvalOptions{
		Step:       step,
		Resolution: candlestick.Interval1d,
		Symbols:    []string{"UNICORN:US:AAPL"},
	})
	bot.SetMaxThreads(1)
	bot.Run(scenarios, paramKeys)
	_, err := json.Marshal(bot.Results())
	if err != nil {
		log.Println("could not save results")
		log.Fatalln(err)
	}
	return bot
}
