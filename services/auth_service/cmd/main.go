package main

import (
	"RegistrationForMessenger/internal/api"
	"RegistrationForMessenger/internal/database"
	"RegistrationForMessenger/internal/logger"
	"RegistrationForMessenger/internal/redi"
	"RegistrationForMessenger/internal/service"
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

func (s *Server) Run(ctx context.Context) {
	s.startTime = time.Now()
	s.router.Use(s.handler.CORS)
	s.router.Use(s.handler.LoggerMiddleware)
	s.router.Post("/register", s.handler.Register)
	//register handlers
	s.serverLogger.Info("log",
		zap.String("status", "started"),
		zap.Time("start_time", time.Now()),
	)
	if err := s.server.ListenAndServe(); err != nil {
		//log
		panic(err)
	}
}

func (s *Server) Close() {
	ctxTime, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	if err := s.server.Shutdown(ctxTime); err != nil {
		statusField := zap.String("status", "stopped_badly")
		if errors.Is(err, http.ErrServerClosed) {
			statusField = zap.String("status", "stopped_gracefully")
		}
		s.serverLogger.Info("log",
			statusField,
			zap.Duration("work_duration", time.Since(s.startTime)),
		)
	}
	//close everything
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := NewServer(ctx)

	<-ctx.Done()

	server.Close()

}
