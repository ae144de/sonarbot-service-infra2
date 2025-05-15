package handler

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter sets up the API routes for the gateway.
func NewRouter() http.Handler {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Stream analysis endpoint (POST)
	router.HandleFunc("/streamanalysis", StreamAnalysisHandler).Methods("POST")

	return router
}

// StreamAnalysisHandler handles POST /streamanalysis requests.
func StreamAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: parse request body, validate, and forward to appropriate service
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"processing"}`))
}
