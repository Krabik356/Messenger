package api

import (
	models2 "chat_manager_service/internal/models"
	"chat_manager_service/internal/service"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	service    *service.Service
	httpLogger *zap.Logger
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateChat(w http.ResponseWriter, r *http.Request) {
	var creationData models2.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&creationData); err != nil {
		http.Error(w, models2.InvalidData.Error(), 400)
		return
	}

	if err := h.service.CreateChat(r.Context(), r.Context().Value("userID").(int), creationData.UserId); err != nil {
		switch err {
		case models2.ServersError:
			http.Error(w, err.Error(), 500)
		default:
			http.Error(w, err.Error(), 409)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	if err := json.NewEncoder(w).Encode(models2.CreateChatReturn{
		Status: "success",
	}); err != nil {
		http.Error(w, models2.ServersError.Error(), 500)
		return
	}
}
