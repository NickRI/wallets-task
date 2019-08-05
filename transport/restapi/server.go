package restapi

import (
	"context"
	"log"
	"net/http"
	"time"
)

const shutdownTimeout = 5 * time.Second

type Server struct {
	http.Server
}

func NewServer(addr string, router http.Handler) *Server {
	return &Server{
		Server: http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

func (o *Server) Run() *Server {
	go func() {
		if err := o.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start listening ", o.Addr, ": ", err)
		}
	}()

	return o
}

func (o *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := o.Server.Shutdown(ctx); err != nil {
		log.Println("failed to shutdown HTTP server")
	}
}
