package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BitofferHub/gateway/internal/config"
	"github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func TestNormalizeHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "http error preserved",
			err:        &middleware.HTTPError{Status: http.StatusUnauthorized, Message: "no authentication"},
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "no authentication",
		},
		{
			name:       "grpc invalid argument",
			err:        grpcstatus.Error(codes.InvalidArgument, "bad request"),
			wantStatus: http.StatusBadRequest,
			wantMsg:    "invalid request",
		},
		{
			name:       "grpc unavailable",
			err:        grpcstatus.Error(codes.Unavailable, "db down"),
			wantStatus: http.StatusServiceUnavailable,
			wantMsg:    "service unavailable",
		},
		{
			name:       "generic error hidden",
			err:        errors.New("sql: connection refused"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeHTTPError(tt.err)
			if got.Status != tt.wantStatus || got.Message != tt.wantMsg {
				t.Fatalf("unexpected http error: got (%d, %q) want (%d, %q)", got.Status, got.Message, tt.wantStatus, tt.wantMsg)
			}
		})
	}
}

func TestWriteErrorHidesInternalDetails(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/bitstorm/get_user_info", nil)
	recorder := httptest.NewRecorder()

	writeError(req, recorder, errors.New("dial tcp 127.0.0.1:3307: connection refused"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: got %d want %d", recorder.Code, http.StatusInternalServerError)
	}
	body := recorder.Body.String()
	if strings.Contains(body, "connection refused") {
		t.Fatalf("unexpected internal error leakage: %s", body)
	}
	if !strings.Contains(body, "internal server error") {
		t.Fatalf("expected generic internal error message, got %s", body)
	}
}

func TestReadyHandlerReturnsServiceUnavailableWhenDependenciesMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	recorder := httptest.NewRecorder()
	svcCtx := &svc.ServiceContext{
		Config: config.Config{},
	}

	ReadyHandler(svcCtx).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: got %d want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "\"status\":\"not_ready\"") {
		t.Fatalf("expected not_ready body, got %s", body)
	}
	if !strings.Contains(body, "\"name\":\"redis\"") {
		t.Fatalf("expected redis check in body, got %s", body)
	}
}
