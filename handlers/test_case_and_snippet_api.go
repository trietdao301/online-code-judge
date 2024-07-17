package handlers

import (
	"encoding/json"
	"example/server/handlers/models"
	"net/http"
)

func (s *apiServerHandler) handleProblemTestCaseAndSubmissionSnippet(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "POST" {
		return s.CreateTestCaseAndSubmissionSnippet(w, r)
	}
	return nil
}
func (s *apiServerHandler) CreateTestCaseAndSubmissionSnippet(w http.ResponseWriter, r *http.Request) error {
	var (
		req     models.CreateTestCaseAndSubmissionSnippetRequest
		context = r.Context()
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

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	err = s.testCaseAndSubmissionSnippetLogic.CreateTestCaseAndSubmissionSnippet(context, &req)
	if err != nil {
		s.logger.Error("fail to create test case and submission snippet")
		return WriteJSON(w, http.StatusBadRequest, "Failed to create test case and submission snippet: "+err.Error())
	}
	return WriteJSON(w, http.StatusOK, "Successfully create test case and submission snippet")
}
