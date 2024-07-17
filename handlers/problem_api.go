package handlers

import (
	"encoding/json"
	"example/server/handlers/models"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
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
		ctx               = r.Context()
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

	params := mux.Vars(r)
	uuid := params["problemUUID"]
	if uuid == "" {
		return WriteJSON(w, http.StatusBadRequest, "Missing UUID parameter")
	}
	getProblemRequest.UUID = uuid
	problem, err := s.problemLogic.GetProblemByUUID(ctx, &models.GetProblemRequest{UUID: getProblemRequest.UUID})
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

	err = json.NewDecoder(r.Body).Decode(&createProblemRequest)
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
	var (
		req models.DeleteProblemRequest
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
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	case RoleAdmin, RoleProblemSetter:
		break
	default:
		return WriteJSON(w, http.StatusUnauthorized, "Insufficient permissions")
	}

	params := mux.Vars(r)
	uuid := params["problemUUID"]
	if uuid == "" {
		return WriteJSON(w, http.StatusBadRequest, "Missing UUID parameter")
	}
	req.ProblemUUID = uuid
	err = s.problemLogic.DeleteProblem(ctx, &req)
	if err != nil {
		s.logger.Error("fail to delete a problem", zap.Any("problemUUID", uuid))
		return WriteJSON(w, http.StatusBadRequest, err)
	}
	return nil
}

func (s *apiServerHandler) UpdateProblem(w http.ResponseWriter, r *http.Request) error {

	return nil
}
