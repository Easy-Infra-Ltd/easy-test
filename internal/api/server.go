package api

import (
	"log/slog"
	"net/http"
)

type Server struct {
	name     string
	addr     string
	servemux *http.ServeMux
	logger   *slog.Logger
}

func NewServer(name string, addr string) *Server {
	logger := slog.Default().With("area", "Server "+name)

	servemux := http.NewServeMux()
	return &Server{
		name:     name,
		addr:     addr,
		servemux: servemux,
		logger:   logger,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(s.addr, s.servemux)
}

func (s *Server) AddRoute(pattern string, handler http.HandlerFunc) {
	s.servemux.Handle(pattern, handler)
}
