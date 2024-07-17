package handlers

import (
	"encoding/json"
	"example/server/handlers/models"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func (s *apiServerHandler) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.GetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.CreateAccount(w, r)
	}
	if r.Method == "DELETE" {
		return s.DeleteAccount(w, r)
	}
	if r.Method == "PUT" {
		return s.UpdateAccount(w, r)
	}
	return nil
}

func (s *apiServerHandler) GetAccount(w http.ResponseWriter, r *http.Request) error {
	var (
		getAccountRequest models.GetAccountRequest
		context           = r.Context()
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
	uuid := params["accountUUID"]
	if uuid == "" {
		return WriteJSON(w, http.StatusBadRequest, "Missing UUID parameter")
	}
	getAccountRequest.UUID = uuid
	account, err := s.accountLogic.GetAccountByUUID(context, &models.GetAccountRequest{UUID: getAccountRequest.UUID})
	if err != nil {
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	return WriteJSON(w, http.StatusOK, account)
}

func (s *apiServerHandler) CreateAccount(w http.ResponseWriter, r *http.Request) error {
	var (
		createAccountRequest models.CreateAccountRequest
		context              = r.Context()
	)
	err := json.NewDecoder(r.Body).Decode(&createAccountRequest)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "Invalid request body")
	}

	res, err := s.accountLogic.CreateAccount(context, &createAccountRequest)
	if err != nil {
		s.logger.Info("Account creating fails")
		return WriteJSON(w, http.StatusOK, err.Error())
	}

	return WriteJSON(w, http.StatusOK, "Successfully created account for "+res.Username)
}

func (s *apiServerHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) error {
	var (
		req models.DeleteAccountRequest
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

	params := mux.Vars(r)
	uuid := params["accountUUID"]
	if uuid == "" {
		return WriteJSON(w, http.StatusBadRequest, "Missing UUID parameter")
	}
	req.UUID = uuid
	err = s.accountLogic.DeleteAccount(ctx, &req)
	if err != nil {
		s.logger.Error("fail to delete an account", zap.Any("accountUUID", uuid))
		return WriteJSON(w, http.StatusInternalServerError, err)
	}
	return WriteJSON(w, http.StatusOK, "Account successfully deleted")
}

func (s *apiServerHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) error {
	var (
		updateAccountRequest models.UpdateAccountRequest
		context              = r.Context()
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
	case RoleProblemSetter:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	case RoleAdmin, RoleContestant:
		break
	default:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	}

	err = json.NewDecoder(r.Body).Decode(&updateAccountRequest)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "Invalid request body")
	}

	res, err := s.accountLogic.UpdateAccount(context, &updateAccountRequest)
	if err != nil {
		s.logger.Error("fail to update an account", zap.Any("accountUUID", updateAccountRequest.UUID))
		return WriteJSON(w, http.StatusInternalServerError, err)
	}

	return WriteJSON(w, http.StatusOK, res)
}
