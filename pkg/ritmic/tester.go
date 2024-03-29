package ritmic

import (
	"encoding/json"
	"github.com/godoji/algocore/internal/simulation"
	"github.com/northberg/candlestick"
	"log"
)

func RunTestShort(step simulation.StepFunction, scenarios [][]float64, paramKeys []string) *simulation.Evaluator {
	bot := simulation.NewEvaluator(simulation.EvalOptions{
		Step:       step,
		Resolution: candlestick.Interval1d,
		Symbols:    []string{"UNICORN:US:COKE"},
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
