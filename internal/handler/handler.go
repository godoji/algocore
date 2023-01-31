package handler

import (
	"encoding/json"
	"fmt"
	"github.com/godoji/algocore/pkg/algo"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
)

type BotStatus struct {
	Running bool
	Model   *Evaluator
}

type EvalConfig struct {
	Symbols    []string    `json:"symbols"`
	Scenarios  [][]float64 `json:"scenarios"`
	Resolution int64       `json:"resolution"`
}

var runLock = sync.Mutex{}
var status = BotStatus{}
var s *Strategy

func handleSimRun(w http.ResponseWriter, r *http.Request) {

	runLock.Lock()
	defer runLock.Unlock()

	// do not run bot twice
	if status.Running {
		http.Error(w, "bot is running", http.StatusServiceUnavailable)
		return
	}

	// check if body is empty
	if r.Body == nil {
		http.Error(w, "no body", http.StatusBadRequest)
		return
	}

	params := new(EvalConfig)
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		http.Error(w, "could not parse variables", http.StatusBadRequest)
		return
	}

	// validate parameters
	if len(params.Symbols) < 1 {
		http.Error(w, "there must be at least 1 symbol", http.StatusBadRequest)
		return
	}

	if params.Resolution == 0 {
		http.Error(w, "invalid resolution", http.StatusBadRequest)
		return
	}

	// create bot
	bot := NewEvaluator(EvalOptions{
		Step:       s.Evaluator,
		Resolution: params.Resolution,
		Symbols:    params.Symbols,
	})

	// enter settings
	bot.SetMaxThreads(12)

	// run bot
	status.Running = true
	status.Model = bot
	go bot.Run(params.Scenarios, s.ParamKeys)

	w.WriteHeader(http.StatusOK)
}

func handleSimStatus(w http.ResponseWriter, _ *http.Request) {
	if status.Model == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status.Model.Metrics())
}

func handleSimResults(w http.ResponseWriter, _ *http.Request) {
	if status.Model == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(status.Model.Results())
	if err != nil {
		fmt.Println(err)
	}
}

func handleHeartbeat(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

type Strategy struct {
	Evaluator algo.StepFunction
	ParamKeys []string
}

func Router(strategy *Strategy) *mux.Router {
	s = strategy
	r := mux.NewRouter()
	r.HandleFunc("/run", handleSimRun).Methods("POST")
	r.HandleFunc("/status", handleSimStatus).Methods("GET")
	r.HandleFunc("/results", handleSimResults).Methods("GET")
	r.HandleFunc("/heartbeat", handleHeartbeat).Methods("GET")
	return r
}
