package handlers

import (
	"encoding/json"
	"example/server/handlers/models"
	"net/http"

	"go.uber.org/zap"
)

func (s *apiServerHandler) handleSession(w http.ResponseWriter, r *http.Request) error {

	if r.Method == "POST" {
		return s.CreateSession(w, r)
	}
	return nil
}

func (s *apiServerHandler) CreateSession(w http.ResponseWriter, r *http.Request) error {
	var (
		loginRequest models.CreateSessionRequest
		ctx          = r.Context()
	)

	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	if err != nil {
		return WriteJSON(w, http.StatusBadRequest, "Invalid request body")
	}

	// Assuming you've added a Login method to your accountLogic
	sessionResponse, err := s.accountLogic.CreateSession(ctx, &loginRequest)
	if err != nil {
		s.logger.Error("failed to login", zap.Error(err), zap.String("username", loginRequest.Username))
		return WriteJSON(w, http.StatusUnauthorized, "Invalid credentials")
	}
	return WriteJSON(w, http.StatusOK, sessionResponse)
}
