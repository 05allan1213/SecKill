package health

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	projectlog "github.com/BitofferHub/seckill/internal/log"
)

type Check struct {
	Name string
	Run  func(context.Context) error
}

type Server struct {
	srv     *http.Server
	service string
	checks  []Check
}

type checkResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type response struct {
	Status    string        `json:"status"`
	Service   string        `json:"service"`
	Checks    []checkResult `json:"checks,omitempty"`
	Timestamp string        `json:"timestamp"`
}

func NewServer(host string, port int, service string, checks []Check) *Server {
	s := &Server{
		service: service,
		checks:  checks,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)
	s.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}
	return s
}

func (s *Server) Start() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			projectlog.Error(nil, "seckill health server stopped unexpectedly",
				projectlog.Field(projectlog.FieldAction, "health.server"),
				projectlog.Field(projectlog.FieldError, err.Error()),
			)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s == nil || s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, response{
		Status:    "ok",
		Service:   s.service,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	resp := response{
		Status:    "ok",
		Service:   s.service,
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    make([]checkResult, 0, len(s.checks)),
	}
	statusCode := http.StatusOK

	for _, check := range s.checks {
		checkCtx, cancel := context.WithTimeout(r.Context(), time.Second)
		err := check.Run(checkCtx)
		cancel()

		result := checkResult{Name: check.Name, Status: "ok"}
		if err != nil {
			statusCode = http.StatusServiceUnavailable
			resp.Status = "not_ready"
			result.Status = "failed"
			result.Error = err.Error()
		}
		resp.Checks = append(resp.Checks, result)
	}

	writeJSON(w, statusCode, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
