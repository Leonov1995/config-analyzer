package http

import (
	"log"
	"net/http"

	"config-analyzer/internal/server"
)

type Server struct {
	handler http.Handler
}

func NewServer(handler http.Handler) *Server {
	return &Server{handler: handler}
}

func (s *Server) Run(addr string) error {
	httpServer := &http.Server{
		Addr:    addr,
		Handler: s.handler,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("REST server started on %s", addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-server.WaitSignalChan():
	}

	return server.GracefulShutdown("REST", httpServer.Shutdown)
}
