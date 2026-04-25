package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	server *http.Server
	router *chi.Mux
	ctx    context.Context
}

func NewServer(ctx context.Context) *Server {
	router := chi.NewRouter()
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	return &Server{
		server: server,
		router: router,
		ctx:    ctx,
	}
}

func (s *Server) Run(ctx context.Context) {
	//register handlers
	if err := s.server.ListenAndServe(); err != nil {
		//log
		panic(err)
	}
}

func (s *Server) Close() {
	ctxTime, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	if err := s.server.Shutdown(ctxTime); err != nil {
		//log
	}
	//close everything
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := NewServer(ctx)

	//log about start
	<-ctx.Done()
	//log about stop

	server.Close()

}
