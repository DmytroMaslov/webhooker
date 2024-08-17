package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

type Server struct {
	port   int
	routes http.Handler
	server *http.Server
}

func NewHttpServer(port int, routes http.Handler) *Server {
	return &Server{
		routes: routes,
		port:   port,
	}

}

func (s *Server) Serve() error {
	// start server
	p := strconv.Itoa(s.port)
	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%s", p),
		Handler: s.routes}

	err := s.server.ListenAndServe()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
