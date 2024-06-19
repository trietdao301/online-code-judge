package handlers

import (
	"encoding/json"
	"example/server/handlers/models"
	"net/http"
)

func (s *apiServerHandler) handleSubmission(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetSubmission(w, r)
	}
	if r.Method == "POST" {
		return s.CreateSubmission(w, r)
	}
	if r.Method == "DELETE" {
		return s.DeleteSubmission(w, r)
	}
	if r.Method == "PUT" {
		return s.UpdateSubmission(w, r)
	}
	return nil
}

func (s *apiServerHandler) GetSubmission(w http.ResponseWriter, r *http.Request) error {
	s.logger.Info("Request GET submission")
	var (
		getSubmissionRequest models.GetSubmissionRequest
		context              = r.Context()
	)
	err := json.NewDecoder(r.Body).Decode(&getSubmissionRequest)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	res, err := s.submissionLogic.GetSubmission(context, &getSubmissionRequest)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	s.logger.Info("Response GET submission")
	return WriteJSON(w, http.StatusOK, res)
}

func (s *apiServerHandler) CreateSubmission(w http.ResponseWriter, r *http.Request) error {
	s.logger.Info("Request POST submission")
	var (
		submissionRequest models.CreateSubmissionRequest
		context           = r.Context()
	)
	// Decode the request body into the Submission struct
	err := json.NewDecoder(r.Body).Decode(&submissionRequest)
	if err != nil {
		s.logger.Error("fail to decode submission request body to submission struct")
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	s.logger.Info("Successfully decode submission request")
	response, err := s.submissionLogic.CreateSubmission(context, &submissionRequest)
	if err != nil {
		s.logger.Error("fail to create submission")
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	s.logger.Info("Response POST submission")
	return WriteJSON(w, http.StatusOK, response)
}

func (s *apiServerHandler) DeleteSubmission(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *apiServerHandler) UpdateSubmission(w http.ResponseWriter, r *http.Request) error {
	return nil
}
