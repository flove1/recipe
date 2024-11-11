package http

import (
	"context"
	"flove/job/config"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config *config.Config
	server *http.Server
	notify chan error
}

func NewServer(cfg *config.Config, handlers Handlers) *Server {
	gin.SetMode(gin.DebugMode)

	router := newRouter(cfg, handlers)

	return &Server{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.HttpPort),
			Handler: router.Handler(),
		},
		config: cfg,
		notify: make(chan error, 1),
	}
}

func (s *Server) Start() {
	go func() {
		s.notify <- s.server.ListenAndServe()
		close(s.notify)
	}()
}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*1000)
	defer cancel()

	return s.server.Shutdown(ctx)
}

func (s *Server) Notify() <-chan error {
	return s.notify
}
