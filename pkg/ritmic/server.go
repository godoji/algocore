package ritmic

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godoji/algocore/internal/simulation"
	"log"
	"net/http"
	"os"
	"strings"
)

type strategy struct {
	Evaluator simulation.StepFunction
	ParamKeys []string
}

var srv *http.Server
var s *strategy

func Serve(evaluate simulation.StepFunction, params []string) {

	s = &strategy{
		Evaluator: evaluate,
		ParamKeys: params,
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8071"
	}
	srv = &http.Server{
		Addr:    ":" + port,
		Handler: router(),
	}
	fmt.Printf("Listening on port %s\n", port)
	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}
}

func sendAsJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}

func sendAsBinary(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)
	_ = gob.NewEncoder(w).Encode(data)
}

func sendResponse(w http.ResponseWriter, r *http.Request, data interface{}) {

	// try to satisfy accept header
	accepts := r.Header.Get("Accept")

	// send as json when nothing is specified
	if accepts == "" {
		sendAsJSON(w, data)
		return
	}

	// send as json when json is requested
	if strings.Index(accepts, "application/json") != -1 {
		sendAsJSON(w, data)
		return
	}

	// send as gob when binary is requested
	if strings.Index(accepts, "application/octet-stream") != -1 {
		sendAsBinary(w, data)
		return
	}

	// send as json when any is requested
	if strings.Index(accepts, "*/*") != -1 {
		sendAsJSON(w, data)
		return
	}

	// deny other types
	w.WriteHeader(http.StatusNotAcceptable)

}
