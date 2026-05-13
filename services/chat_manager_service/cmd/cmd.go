package main

import (
	"chat_manager_service/config"
	"chat_manager_service/internal/api"
	"chat_manager_service/internal/database"
	"chat_manager_service/internal/kafk"
	"chat_manager_service/internal/logger"
	"chat_manager_service/internal/service"
	"chat_manager_service/internal/token"
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
	handler      *api.Handler
	router       *chi.Mux
	service      *service.Service
	serverLogger *zap.Logger
	consumer     *kafk.Consumer
	producer     *kafk.Producer
	startTime    time.Time
}

func NewServer(ctx context.Context) *Server {
	config := config.NewConfig()
	logger := logger.NewLogger()
	router := chi.NewRouter()
	token := token.NewToken(config.TokensSecret)
	database := database.NewDatabase(config.GetPostgresUrl())
	service := service.NewService(database, token)
	consumer := kafk.NewConsumer(ctx, service, logger.ConsLogger, config.GetConsumerUrl())
	producer := kafk.NewProducer(service)
	handler := api.NewHandler(service, logger.HttpLogger)
	server := &http.Server{
		Addr:    config.GetServerUrl(),
		Handler: router,
	}

	return &Server{
		server:       server,
		handler:      handler,
		router:       router,
		service:      service,
		serverLogger: logger.ServerLogger,
		consumer:     consumer,
		producer:     producer,
	}
}

func (s *Server) Run() {
	s.startTime = time.Now()
	//s.router.Use(s.handler.CORS) ------------------------------------------UNCOMMENT!!!
	s.router.Use(s.handler.Authorization)
	s.router.Post("/create_chat", s.handler.CreateChat)
	s.router.Post("/send_message", s.handler.SendMessage)
	s.router.Get("/get_data", s.handler.GetUsersData)
	s.router.Get("/get_messages/{chat_id}", s.handler.GetChatsMessages)
	go s.producer.Produce()
	go s.producer.EventListener()
	go s.consumer.Consume()

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
	errorField := zap.Field{}
	isGrace := true

	s.service.Close()
	if err := s.consumer.Close(); err != nil {
		statusField = zap.String("status", "stopped_badly")
		errorField = zap.String("error", err.Error())
		isGrace = false
	}

	if err := s.server.Shutdown(ctxTime); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			statusField = zap.String("status", "stopped_badly")
			errorField = zap.String("error", err.Error())
			isGrace = false
		}
	}

	fields := []zap.Field{
		statusField,
		zap.Duration("work_duration", time.Since(s.startTime)),
	}
	if !isGrace {
		fields = append(fields, errorField)
		s.serverLogger.Error("server stopped", fields...)
		return
	}
	s.serverLogger.Info("server stopped", fields...)
	return

}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := NewServer(ctx)
	go server.Run()

	<-ctx.Done()

	server.Close()
}
