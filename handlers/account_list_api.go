package handlers

import (
	"net/http"
)

func (s *apiServerHandler) handleAccountList(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetAccountList(w, r)
	}
	return nil
}

func (s *apiServerHandler) GetAccountList(w http.ResponseWriter, r *http.Request) error {
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
	case RoleContestant, RoleProblemSetter:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	case RoleAdmin:
		break
	default:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	}

	res, err := s.accountLogic.GetAllAccounts(ctx)
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	return WriteJSON(w, http.StatusOK, res)
}
