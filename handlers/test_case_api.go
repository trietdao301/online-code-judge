package handlers

import (
	"encoding/json"
	"example/server/handlers/models"
	"net/http"
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
	err := json.NewDecoder(r.Body).Decode(&getTestCaseRequest)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	testCase, err := s.testCaseLogic.GetTestCaseByUUID(context, &models.GetTestCaseRequest{UUID: getTestCaseRequest.UUID})
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
	err := json.NewDecoder(r.Body).Decode(&createTestCaseRequest)
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
