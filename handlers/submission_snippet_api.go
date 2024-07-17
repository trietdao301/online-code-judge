package handlers

import (
	"encoding/json"
	"example/server/handlers/models"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (s *apiServerHandler) handleSubmissionSnippet(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetSubmissionSnippet(w, r)
	}
	if r.Method == "POST" {
		return s.CreateSubmissionSnippet(w, r)
	}
	return nil
}

func (s *apiServerHandler) GetSubmissionSnippet(w http.ResponseWriter, r *http.Request) error {
	s.logger.Info("Request GET submission snippet")
	var (
		getSubmissionSnippetRequest models.GetSubmissionSnippetRequest
		context                     = r.Context()
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
	case RoleContestant, RoleAdmin, RoleProblemSetter:
		break
	default:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	}

	params := mux.Vars(r)
	uuid := params["submissionSnippetUUID"]
	if uuid == "" {
		return WriteJSON(w, http.StatusBadRequest, "Missing UUID parameter")
	}
	getSubmissionSnippetRequest.SubmissionSnippetUUID = uuid

	res, err := s.submissionSnippetLogic.GetSubmissionSnippetByUUID(context, &getSubmissionSnippetRequest)
	if err != nil {
		s.logger.Error("fail to get submission snippet by uuid", zap.Any("snippet uuid", getSubmissionSnippetRequest.SubmissionSnippetUUID))
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	return WriteJSON(w, http.StatusOK, res)
}

func (s *apiServerHandler) CreateSubmissionSnippet(w http.ResponseWriter, r *http.Request) error {
	var (
		createSubmissionSnippetRequest models.CreateSubmissionSnippetRequest
		context                        = r.Context()
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

	err = json.NewDecoder(r.Body).Decode(&createSubmissionSnippetRequest)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	err = s.submissionSnippetLogic.CreateSubmissionSnippet(context, &createSubmissionSnippetRequest)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	return WriteJSON(w, http.StatusOK, "Succefully create a submission snippet")
}
