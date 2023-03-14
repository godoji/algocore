package ritmic

import (
	"fmt"
	"github.com/godoji/algocore/internal/handler"
	"github.com/godoji/algocore/internal/simulation"
	"log"
	"net/http"
	"os"
)

func Serve(evaluate simulation.StepFunction, params []string) {

	strategy := &handler.Strategy{
		Evaluator: evaluate,
		ParamKeys: params,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8071"
	}
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler.Router(strategy),
	}
	fmt.Printf("service has started on port %s\n", port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln(err)
	}
}
