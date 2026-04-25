package api

import (
	"Messenger/internal/models"
	"Messenger/internal/service"
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	service    *service.Service
	httpLogger *zap.Logger
}

func NewHandler(service *service.Service, httpLogger *zap.Logger) *Handler {
	return &Handler{
		service:    service,
		httpLogger: httpLogger,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var regData models.RegRequest
	if err := json.NewDecoder(r.Body).Decode(&regData); err != nil {
		http.Error(w, models.InvalidData.Error(), 400)
		return
	}

	if err := h.service.Register(regData.Name, regData.Email, regData.Password); err != nil {
		switch {
		case errors.Is(err, models.ServersError):
			http.Error(w, err.Error(), 500)
			return
		case errors.Is(err, models.AlreadyExists):
			http.Error(w, err.Error(), 409)
			return
		default:
			http.Error(w, err.Error(), 400)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	if err := json.NewEncoder(w).Encode(models.RegResp{
		Status: "success",
	}); err != nil {
		http.Error(w, models.ServersError.Error(), 500)
		return
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var logData models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&logData); err != nil {
		http.Error(w, models.InvalidData.Error(), 400)
		return
	}

	isExists, err := h.service.Login(logData.Email, logData.Password)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if !isExists {
		http.Error(w, models.UnknownUser.Error(), 404)
		return
	}

	refreshToken, accessToken, err := h.service.GenerateJWTTokens(logData.Email)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	if err := json.NewEncoder(w).Encode(models.LoginResp{
		Status:  "success",
		Refresh: refreshToken,
		Access:  accessToken,
	}); err != nil {
		http.Error(w, models.ServersError.Error(), 500)
		return
	}
}
