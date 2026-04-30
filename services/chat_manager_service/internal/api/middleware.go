package api

import (
	"chat_manager_service/models"
	"context"
	"net/http"
	"strings"
)

type ResponceWriter struct {
	http.ResponseWriter
	code int
}

func NewResponceWriter(w http.ResponseWriter) *ResponceWriter {
	return &ResponceWriter{
		ResponseWriter: w,
		code:           200,
	}
}

func (rw *ResponceWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}

func (h *Handler) Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorString := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authorString) < 2 {
			http.Error(w, models.InvalidToken.Error(), 400)
			return
		}

		if authorString[0] != "Bearer" {
			http.Error(w, models.InvalidToken.Error(), 400)
			return
		}

		id, isValid, err := h.service.IsValidToken(authorString[1])
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if !isValid {
			http.Error(w, models.InvalidToken.Error(), 400)
			return
		}

		rr := r.WithContext(context.WithValue(r.Context(), "userID", id))
		next.ServeHTTP(w, rr)
	})
}
