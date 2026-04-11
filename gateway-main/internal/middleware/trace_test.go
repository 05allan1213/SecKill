package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BitofferHub/pkg/constant"
)

func TestTraceMiddleware_UsesIncomingTraceID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	req.Header.Set(constant.TraceID, "trace-123")
	recorder := httptest.NewRecorder()

	NewTraceMiddleware().Handle(func(w http.ResponseWriter, r *http.Request) {
		if got := TraceIDFromContext(r.Context()); got != "trace-123" {
			t.Fatalf("unexpected trace id in context: %q", got)
		}
	})(recorder, req)

	if got := recorder.Header().Get(constant.TraceID); got != "trace-123" {
		t.Fatalf("unexpected response trace id: %q", got)
	}
}

func TestTraceMiddleware_GeneratesTraceID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	recorder := httptest.NewRecorder()

	NewTraceMiddleware().Handle(func(w http.ResponseWriter, r *http.Request) {
		if got := TraceIDFromContext(r.Context()); got == "" {
			t.Fatal("expected generated trace id")
		}
	})(recorder, req)

	if got := recorder.Header().Get(constant.TraceID); got == "" {
		t.Fatal("expected response trace id")
	}
}
