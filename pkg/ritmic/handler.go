package ritmic

import (
	"context"
	"encoding/json"
	"github.com/godoji/algocore/internal/simulation"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"sync"
	"time"
)

type EvaluateConfig struct {
	Symbols    []string    `json:"symbols"`
	Scenarios  [][]float64 `json:"scenarios"`
	Resolution int64       `json:"resolution"`
}

var wg sync.WaitGroup
var isTerminating = false

func handleTerminate(w http.ResponseWriter, _ *http.Request) {
	isTerminating = true
	wg.Wait()
	w.WriteHeader(http.StatusOK)
	go func() {
		time.Sleep(10 * time.Millisecond)
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatal("Server shutdown failed: ", err)
		}
	}()
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {

	// Add to wait group
	wg.Add(1)
	defer wg.Done()

	// Check if the stop signal has been received
	if isTerminating {
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
	}

	// Check the request body
	if r.Body == nil {
		http.Error(w, "no body", http.StatusBadRequest)
		return
	}

	// Parse request parameters
	params := new(EvaluateConfig)
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		http.Error(w, "could not parse variables", http.StatusBadRequest)
		return
	}
	if len(params.Symbols) < 1 {
		http.Error(w, "there must be at least 1 symbol", http.StatusBadRequest)
		return
	}
	if params.Resolution == 0 {
		http.Error(w, "invalid resolution", http.StatusBadRequest)
		return
	}

	// Create an evaluator to run requested scenario
	evaluator := simulation.NewEvaluator(simulation.EvalOptions{
		Step:       s.Evaluator,
		Resolution: params.Resolution,
		Symbols:    params.Symbols,
	})

	// Run the simulation with given parameters
	evaluator.SetMaxThreads(4)
	evaluator.Run(params.Scenarios, s.ParamKeys)

	// Send back the results as a sync request
	sendResponse(w, r, evaluator.Results())
}

func handleHeartbeat(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/terminate", handleTerminate).Methods("POST")
	r.HandleFunc("/evaluate", handleEvaluate).Methods("POST")
	r.HandleFunc("/heartbeat", handleHeartbeat).Methods("GET")
	return r
}
