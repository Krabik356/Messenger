package main

import (
	"Messenger/internal/api"
	"Messenger/internal/database"
	"Messenger/internal/logger"
	"Messenger/internal/redi"
	"Messenger/internal/service"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Server struct {
	server       *http.Server
	router       *chi.Mux
	database     *database.Database
	redi         *redi.Redis
	service      *service.Service
	handler      *api.Handler
	serverLogger *zap.Logger
	ctx          context.Context
	startTime    time.Time
}

func NewServer(ctx context.Context) *Server {
	logger := logger.NewLogger()
	redi := redi.NewRedis(ctx)
	database := database.NewDatabase(ctx)
	service := service.NewService(ctx, database, redi)
	handler := api.NewHandler(service, logger.HttpLogger)
	router := chi.NewRouter()
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	return &Server{
		server:       server,
		router:       router,
		database:     database,
		redi:         redi,
		service:      service,
		handler:      handler,
		serverLogger: logger.ServerLogger,
		ctx:          ctx,
	}
}

func (s *Server) Run() {
	s.startTime = time.Now()
	//s.router.Use(s.handler.CORS) ------------------------------------------UNCOMMENT!!!
	s.router.Use(s.handler.LoggerMiddleware)
	s.router.Post("/register", s.handler.Register)
	s.router.Post("/login", s.handler.Login)
	s.router.Get("/refresh", s.handler.Refresh)
	//register handlers
	s.serverLogger.Info("server started",
		zap.String("start_time", time.Now().Format(time.RFC3339)),
	)
	if err := s.server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}
}

func (s *Server) Close() {
	ctxTime, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	statusField := zap.String("status", "stopped_gracefully")
	isGrace := true

	if err := s.server.Shutdown(ctxTime); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			statusField = zap.String("status", "stopped_badly")
			isGrace = false
		}
	}
	if err := s.service.Close(); err != nil {
		statusField = zap.String("status", "stopped_badly")
		isGrace = false
	}

	fields := []zap.Field{
		statusField,
		zap.Duration("work_duration", time.Since(s.startTime)),
	}
	if !isGrace {
		s.serverLogger.Error("server stopped", fields...)
		return
	}
	s.serverLogger.Info("server stopped", fields...)
	//close everything
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := NewServer(ctx)
	go server.Run()

	<-ctx.Done()

	server.Close()
}
