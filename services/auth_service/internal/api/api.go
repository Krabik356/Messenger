package api

import (
	"RegistrationForMessenger/internal/models"
	"RegistrationForMessenger/internal/service"
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
		http.Error(w, "invalid data", 400)
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

	refreshToken, accessToken, err := h.service.GenerateJWTTokens(regData.Email)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	if err := json.NewEncoder(w).Encode(models.RegResp{
		Status:  "success",
		Refresh: refreshToken,
		Access:  accessToken,
	}); err != nil {
		http.Error(w, models.ServersError.Error(), 500)
		return
	}
}
