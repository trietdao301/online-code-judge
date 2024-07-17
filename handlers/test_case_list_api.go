package handlers

import (
	"example/server/handlers/models"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *apiServerHandler) handleTestCaseList(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetTestCaseList(w, r)
	}
	return nil
}

func (s *apiServerHandler) GetTestCaseList(w http.ResponseWriter, r *http.Request) error {
	var (
		request models.GetTestCaseListRequest
		ctx     = r.Context()
	)
	token, err := s.validateRequestAndExtractToken(r)
	if err != nil {
		return WriteJSON(w, http.StatusUnauthorized, err.Error())
	}
	_, role, _, err := s.tokenLogic.ExtractTokenData(ctx, token)
	if err != nil {
		return WriteJSON(w, http.StatusUnauthorized, err.Error())
	}

	switch role {
	case RoleContestant:
		// Maybe contestants can only see a subset of test cases
		return WriteJSON(w, http.StatusForbidden, "Insufficient permissions")
	case RoleAdmin, RoleProblemSetter:
		break
	default:
		return WriteJSON(w, http.StatusForbidden, "Insufficient permissions")
	}

	params := mux.Vars(r)
	uuid := params["problemUUID"]
	if uuid == "" {
		return WriteJSON(w, http.StatusBadRequest, "Missing UUID parameter")
	}
	request.ProblemUUID = uuid

	defer r.Body.Close()

	res, err := s.problemLogic.GetAllTestCasesByProblemUUID(ctx, &request)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err.Error())
	}

	return WriteJSON(w, http.StatusOK, res)
}
