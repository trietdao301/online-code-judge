package handlers

import (
	"encoding/json"
	"example/server/handlers/models"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *apiServerHandler) handleTestCase(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetTestCase(w, r)
	}
	if r.Method == "POST" {
		return s.CreateTestCase(w, r)
	}
	if r.Method == "DELETE" {
		return s.DeleteTestCase(w, r)
	}
	if r.Method == "PUT" {
		return s.UpdateTestCase(w, r)
	}
	return nil
}

func (s *apiServerHandler) GetTestCase(w http.ResponseWriter, r *http.Request) error {
	var (
		getTestCaseRequest models.GetTestCaseRequest
		context            = r.Context()
	)

	token, err := s.validateRequestAndExtractToken(r)
	if err != nil {
		return WriteJSON(w, http.StatusUnauthorized, err.Error())
	}
	_, role, _, err := s.tokenLogic.ExtractTokenData(context, token)
	if err != nil {
		return WriteJSON(w, http.StatusUnauthorized, err.Error())
	}
	switch role {
	case RoleContestant:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	case RoleAdmin, RoleProblemSetter:
		break
	default:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	}

	params := mux.Vars(r)
	uuid := params["testUUID"]
	if uuid == "" {
		return WriteJSON(w, http.StatusBadRequest, "Missing UUID parameter")
	}
	getTestCaseRequest.UUID = uuid

	testCase, err := s.testCaseLogic.GetTestCaseByUUID(context, &models.GetTestCaseRequest{UUID: getTestCaseRequest.UUID})
	if testCase == nil {
		return WriteJSON(w, http.StatusInternalServerError, "testCase is nil")
	}
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	return WriteJSON(w, http.StatusOK, testCase)
}

func (s *apiServerHandler) CreateTestCase(w http.ResponseWriter, r *http.Request) error {
	var (
		createTestCaseRequest models.CreateTestCaseRequest
		context               = r.Context()
	)

	token, err := s.validateRequestAndExtractToken(r)
	if err != nil {
		return WriteJSON(w, http.StatusUnauthorized, err.Error())
	}
	_, role, _, err := s.tokenLogic.ExtractTokenData(context, token)
	if err != nil {
		return WriteJSON(w, http.StatusUnauthorized, err.Error())
	}
	switch role {
	case RoleContestant:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	case RoleAdmin, RoleProblemSetter:
		break
	default:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	}

	err = json.NewDecoder(r.Body).Decode(&createTestCaseRequest)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	err = s.testCaseLogic.CreateTestCase(context, &createTestCaseRequest)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	return WriteJSON(w, http.StatusOK, "Succefully create a test case")
}

func (s *apiServerHandler) DeleteTestCase(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *apiServerHandler) UpdateTestCase(w http.ResponseWriter, r *http.Request) error {
	return nil
}
