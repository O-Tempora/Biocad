package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) InitRouter() {
	s.router = chi.NewMux()
}
