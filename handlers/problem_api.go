package handlers

import (
	"encoding/json"
	"example/server/handlers/models"
	"net/http"
)

func (s *apiServerHandler) handleProblem(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetProblem(w, r)
	}
	if r.Method == "POST" {
		return s.CreateProblem(w, r)
	}
	if r.Method == "DELETE" {
		return s.DeleteProblem(w, r)
	}
	if r.Method == "PUT" {
		return s.UpdateProblem(w, r)
	}
	return nil
}

func (s *apiServerHandler) GetProblem(w http.ResponseWriter, r *http.Request) error {
	var (
		getProblemRequest models.GetProblemRequest
		context           = r.Context()
	)
	err := json.NewDecoder(r.Body).Decode(&getProblemRequest)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	problem, err := s.problemLogic.GetProblemByUUID(context, &models.GetProblemRequest{UUID: getProblemRequest.UUID})
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	return WriteJSON(w, http.StatusOK, problem)
}

func (s *apiServerHandler) CreateProblem(w http.ResponseWriter, r *http.Request) error {
	var (
		createProblemRequest models.CreateProblemRequest
		context              = r.Context()
	)
	err := json.NewDecoder(r.Body).Decode(&createProblemRequest)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	res, err := s.problemLogic.CreateProblem(context, &createProblemRequest)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	return WriteJSON(w, http.StatusOK, "Succefully create a problem "+res.DisplayName)
}

func (s *apiServerHandler) DeleteProblem(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *apiServerHandler) UpdateProblem(w http.ResponseWriter, r *http.Request) error {
	return nil
}
