package api

import (
	"RegistrationForMessenger/internal/models"
	"RegistrationForMessenger/internal/service"
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
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

	//JWT

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	if err := json.NewEncoder(w).Encode(models.RegResp{
		Status: "success",
	}); err != nil {
		http.Error(w, models.ServersError.Error(), 500)
		return
	}
}
