package handlers

import (
	"encoding/json"
	"example/server/configs"
	"example/server/logic"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, err.Error())
		}
	}
}

type apiServerHandler struct {
	submissionLogic logic.Submission
	testCaseLogic   logic.TestCase
	config          configs.Config
	logger          *zap.Logger
	problemLogic    logic.Problem
}

func NewAPIServerHandler(submissionLogic logic.Submission, testCaseLogic logic.TestCase, configs configs.Config, problemLogic logic.Problem, logger *zap.Logger) *apiServerHandler {
	return &apiServerHandler{
		submissionLogic: submissionLogic,
		testCaseLogic:   testCaseLogic,
		config:          configs,
		logger:          logger,
		problemLogic:    problemLogic,
	}
}

func (s *apiServerHandler) Start() {
	router := mux.NewRouter()
	address := s.config.Http.Address
	log.Printf("Server started at" + " " + address)

	router.HandleFunc("/submission", makeHTTPHandleFunc(s.handleSubmission))
	router.HandleFunc("/test-case", makeHTTPHandleFunc(s.handleTestCase))
	router.HandleFunc("/problem", makeHTTPHandleFunc(s.handleProblem))
	log.Fatal(http.ListenAndServe(address, router))
}
