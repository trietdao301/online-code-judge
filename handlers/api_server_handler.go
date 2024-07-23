package handlers

import (
	"encoding/json"
	"example/server/configs"
	"example/server/logic"
	"fmt"
	"strings"

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
		//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Origin", "https://online-code-judge-nstj0witj-trietdao301s-projects.vercel.app/")

		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, err.Error())
		}
	}
}

const (
	RoleContestant    = "Contestant"
	RoleAdmin         = "Admin"
	RoleProblemSetter = "ProblemSetter"
)

type apiServerHandler struct {
	submissionLogic                   logic.Submission
	testCaseLogic                     logic.TestCase
	config                            configs.Config
	logger                            *zap.Logger
	problemLogic                      logic.Problem
	submissionSnippetLogic            logic.SubmissionSnippet
	testCaseAndSubmissionSnippetLogic logic.TestCaseAndSubmissionSnippet
	accountLogic                      logic.Account
	tokenLogic                        logic.Token
}

func NewAPIServerHandler(submissionLogic logic.Submission,
	testCaseLogic logic.TestCase,
	configs configs.Config,
	problemLogic logic.Problem,
	submissionSnippetLogic logic.SubmissionSnippet,
	testCaseAndSubmissionSnippet logic.TestCaseAndSubmissionSnippet,
	accountLogic logic.Account,
	tokenLogic logic.Token,
	logger *zap.Logger) *apiServerHandler {
	return &apiServerHandler{
		submissionLogic:                   submissionLogic,
		testCaseLogic:                     testCaseLogic,
		config:                            configs,
		logger:                            logger,
		problemLogic:                      problemLogic,
		submissionSnippetLogic:            submissionSnippetLogic,
		testCaseAndSubmissionSnippetLogic: testCaseAndSubmissionSnippet,
		tokenLogic:                        tokenLogic,
		accountLogic:                      accountLogic,
	}
}

func (s *apiServerHandler) validateRequestAndExtractToken(r *http.Request) (tokenString string, err error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Missing authorization header")
	}

	bearerToken := strings.Split(authHeader, " ")
	if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
		return "", fmt.Errorf("Invalid authorization header format")
	}
	token := bearerToken[1]
	return token, nil
}

func (s *apiServerHandler) Start() {
	router := mux.NewRouter()
	address := s.config.Http.Address
	log.Printf("Server started at" + " " + address)

	// authenticatedRouter := router.PathPrefix("").Subrouter()
	// authenticatedRouter.Use(func(next http.Handler) http.Handler {
	// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		s.authMiddleware(next.ServeHTTP).ServeHTTP(w, r)
	// 	})
	// })

	router.HandleFunc("/submission", makeHTTPHandleFunc(s.handleSubmission))
	router.HandleFunc("/submission/{submissionUUID}", makeHTTPHandleFunc(s.handleSubmission))
	router.HandleFunc("/submission-list/{problemUUID}/{authorAccountUUID}", makeHTTPHandleFunc(s.handleSubmissionList))
	router.HandleFunc("/test-case/{testUUID}", makeHTTPHandleFunc(s.handleTestCase))
	router.HandleFunc("/test-case-list/{problemUUID}", makeHTTPHandleFunc(s.handleTestCaseList))
	router.HandleFunc("/test-case", makeHTTPHandleFunc(s.handleTestCase))
	router.HandleFunc("/problem/{problemUUID}", makeHTTPHandleFunc(s.handleProblem))
	router.HandleFunc("/test-case-and-submission-snippet", makeHTTPHandleFunc(s.handleProblemTestCaseAndSubmissionSnippet))
	router.HandleFunc("/problem", makeHTTPHandleFunc(s.handleProblem))
	router.HandleFunc("/problem-list", makeHTTPHandleFunc(s.handleProblemList))
	router.HandleFunc("/submission-snippet/{submissionSnippetUUID}", makeHTTPHandleFunc(s.handleSubmissionSnippet))
	router.HandleFunc("/submission-snippet", makeHTTPHandleFunc(s.handleSubmissionSnippet))

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{accountUUID}", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/account-list", makeHTTPHandleFunc(s.handleAccountList))
	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleSession))
	log.Fatal(http.ListenAndServe(address, router))
}
