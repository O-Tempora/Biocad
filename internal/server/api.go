package server

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"log/slog"

	"github.com/go-chi/chi/v5"
)

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}, err error) {
	w.WriteHeader(code)
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		s.logger.LogAttrs(context.Background(), slog.LevelError, "Response with error:",
			slog.String("URL", r.URL.Path),
			slog.String("Method", r.Method),
			slog.Int("HTTP Code", code),
			slog.String("HTTP Status", http.StatusText(code)),
			slog.String("Error", err.Error()),
		)
		return
	}

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
	s.logger.LogAttrs(context.Background(), slog.LevelInfo, "Response:",
		slog.String("URL", r.URL.Path),
		slog.String("Method", r.Method),
		slog.Int("HTTP Code", code),
		slog.String("HTTP Status", http.StatusText(code)),
	)
}

func (s *server) InitRouter() {
	s.router = chi.NewMux()
	s.router.Get("/docs", s.handleGetDocs)
}

func (s *server) handleGetDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	files, err := s.service.GetDocs(page, limit)
	if err != nil {
		s.respond(w, r, http.StatusInternalServerError, nil, err)
		return
	}
	s.respond(w, r, http.StatusOK, files, nil)
}
