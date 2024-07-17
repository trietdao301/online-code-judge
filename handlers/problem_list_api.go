package handlers

import (
	"net/http"
)

func (s *apiServerHandler) handleProblemList(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetProblemList(w, r)
	}
	return nil
}

func (s *apiServerHandler) GetProblemList(w http.ResponseWriter, r *http.Request) error {
	var (
		ctx = r.Context()
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
		break
	case RoleAdmin, RoleProblemSetter:
		break
	default:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	}
	res, err := s.problemLogic.GetAllProblems(ctx)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	return WriteJSON(w, http.StatusOK, res)
}
