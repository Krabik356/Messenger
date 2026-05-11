package api

import (
	"chat_manager_service/internal/models"
	"chat_manager_service/internal/service"
	"encoding/json"
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

func (h *Handler) CreateChat(w http.ResponseWriter, r *http.Request) {
	var creationData models.CreateChatRequest
	if err := json.NewDecoder(r.Body).Decode(&creationData); err != nil {
		http.Error(w, models.InvalidData.Error(), 400)
		return
	}

	if err := h.service.CreateChat(r.Context(), r.Context().Value("userID").(int), creationData.UserId, creationData.ChatName); err != nil {
		switch err {
		case models.ServersError:
			http.Error(w, err.Error(), 500)
		default:
			http.Error(w, err.Error(), 409)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	if err := json.NewEncoder(w).Encode(models.CreateChatReturn{
		Status: "success",
	}); err != nil {
		http.Error(w, models.ServersError.Error(), 500)
		return
	}
}

func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var sendData models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&sendData); err != nil {
		http.Error(w, models.InvalidData.Error(), 400)
		return
	}

	id, err := h.service.SendMessage(r.Context(), sendData.ChatId, sendData.Id, sendData.Message)
	if err != nil {
		switch err {
		case models.ServersError:
			http.Error(w, err.Error(), 500)
		case models.NoUserInChat:
			http.Error(w, err.Error(), 404)
		default:
			http.Error(w, models.ServersError.Error(), 500)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	if err := json.NewEncoder(w).Encode(models.SendMessageReturn{
		Status:    "success",
		MessageId: id,
	}); err != nil {
		http.Error(w, models.ServersError.Error(), 500)
		return
	}
}

func (h *Handler) GetUsersData(w http.ResponseWriter, r *http.Request) {
	
	userid, ok := r.Context().Value("id").(int)
	if !ok {
		http.Error(w, models.ServersError.Error(), 404)
		return
	}

	data, err := h.service.GetUsersData(r.Context(), userid)
	if err != nil {
		switch err {
		case models.NoUserInChat:
			http.Error(w, err.Error(), 409)
		default:
			http.Error(w, err.Error(), 500)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

}
