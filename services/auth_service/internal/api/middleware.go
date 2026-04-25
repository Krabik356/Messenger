package api

import (
	"net/http"
	"time"

	"go.uber.org/zap"
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

func (h *Handler) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:7010")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, session_token")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := NewResponceWriter(w)
		startTime := time.Now()
		defer func() {
			code := rw.code
			fields := []zap.Field{
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.Int("status_code", code),
				zap.Duration("duration", time.Since(startTime)),
			}

			if code >= 200 && code < 300 {
				h.httpLogger.Info("log", fields...)
			} else if code >= 400 && code < 500 {
				h.httpLogger.Warn("log", fields...)
			} else if code >= 500 && code < 600 {
				h.httpLogger.Error("log", fields...)
			}
		}()

		next.ServeHTTP(rw, r)
	})
}
