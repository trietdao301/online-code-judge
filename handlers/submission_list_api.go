package handlers

import (
	"example/server/handlers/models"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *apiServerHandler) handleSubmissionList(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetSubmissionList(w, r)
	}
	return nil
}

func (s *apiServerHandler) GetSubmissionList(w http.ResponseWriter, r *http.Request) error {
	var (
		submissions *models.GetSubmissionListResponse
		ctx         = r.Context()
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
	case RoleContestant, RoleAdmin:
		break
	case RoleProblemSetter:
		return WriteJSON(w, http.StatusUnauthorized, "Problem Setter can't get submission")
	default:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	}

	params := mux.Vars(r)
	problemUUID := params["problemUUID"]
	if problemUUID == "" {
		return WriteJSON(w, http.StatusBadRequest, "Missing problemUUID parameter")
	}

	authorAccountUUID := params["authorAccountUUID"]
	if authorAccountUUID == "" {
		return WriteJSON(w, http.StatusBadRequest, "Missing authorAccountUUID parameter")
	}

	submissions, err = s.submissionLogic.GetSubmissionsByProblemAndAuthor(ctx, problemUUID, authorAccountUUID)

	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, "Failed to retrieve submissions")
	}

	return WriteJSON(w, http.StatusOK, submissions)
}
