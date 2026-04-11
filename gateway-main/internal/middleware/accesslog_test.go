package middleware

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/BitofferHub/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
)

type logEntry struct {
	level  string
	msg    interface{}
	fields map[string]interface{}
}

type captureWriter struct {
	entries []logEntry
}

func (w *captureWriter) Alert(v any)                         {}
func (w *captureWriter) Severe(v any)                        {}
func (w *captureWriter) Stack(v any)                         {}
func (w *captureWriter) Stat(v any, fields ...logx.LogField) {}
func (w *captureWriter) Close() error                        { return nil }

func (w *captureWriter) Debug(v any, fields ...logx.LogField) {
	w.entries = append(w.entries, newEntry("debug", v, fields...))
}

func (w *captureWriter) Error(v any, fields ...logx.LogField) {
	w.entries = append(w.entries, newEntry("error", v, fields...))
}

func (w *captureWriter) Info(v any, fields ...logx.LogField) {
	w.entries = append(w.entries, newEntry("info", v, fields...))
}

func (w *captureWriter) Slow(v any, fields ...logx.LogField) {
	w.entries = append(w.entries, newEntry("slow", v, fields...))
}

func newEntry(level string, msg any, fields ...logx.LogField) logEntry {
	entry := logEntry{
		level:  level,
		msg:    msg,
		fields: make(map[string]interface{}, len(fields)),
	}
	for _, field := range fields {
		entry.fields[field.Key] = field.Value
	}
	return entry
}

func TestAccessLogMiddleware_Success(t *testing.T) {
	writer := installTestWriter()
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"username":"admin"}`))
	req = req.WithContext(WithUserID(WithTraceID(context.Background(), "trace-success"), "42"))
	recorder := httptest.NewRecorder()

	middleware := NewAccessLogMiddleware()
	handler := middleware.Handle(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body failed: %v", err)
		}
		if string(body) != `{"username":"admin"}` {
			t.Fatalf("unexpected request body: %q", string(body))
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	handler(recorder, req)

	if len(writer.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(writer.entries))
	}
	entry := writer.entries[0]
	if entry.level != "info" {
		t.Fatalf("expected info level, got %s", entry.level)
	}
	if entry.msg != "http request completed" {
		t.Fatalf("unexpected message: %#v", entry.msg)
	}
	if entry.fields[constant.TraceID] != "trace-success" {
		t.Fatalf("unexpected trace field: %#v", entry.fields[constant.TraceID])
	}
	if entry.fields["status"] != http.StatusCreated {
		t.Fatalf("unexpected status: %#v", entry.fields["status"])
	}
	if entry.fields["user_id"] != "42" {
		t.Fatalf("unexpected user id: %#v", entry.fields["user_id"])
	}
	if entry.fields["reply_summary"] != `{"ok":true}` {
		t.Fatalf("unexpected reply summary: %#v", entry.fields["reply_summary"])
	}
}

func TestAccessLogMiddleware_Error(t *testing.T) {
	writer := installTestWriter()
	req := httptest.NewRequest(http.MethodGet, "/get_user_info?foo=bar", nil)
	req = req.WithContext(WithTraceID(context.Background(), "trace-error"))
	recorder := httptest.NewRecorder()

	middleware := NewAccessLogMiddleware()
	handler := middleware.Handle(func(w http.ResponseWriter, r *http.Request) {
		WriteCodeMessage(w, http.StatusUnauthorized, http.StatusUnauthorized, "no authentication")
	})
	handler(recorder, req)

	if len(writer.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(writer.entries))
	}
	entry := writer.entries[0]
	if entry.level != "error" {
		t.Fatalf("expected error level, got %s", entry.level)
	}
	if entry.msg != "http request failed" {
		t.Fatalf("unexpected message: %#v", entry.msg)
	}
	if entry.fields["status"] != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %#v", entry.fields["status"])
	}
	if entry.fields["req_summary"] != "foo=bar" {
		t.Fatalf("unexpected req summary: %#v", entry.fields["req_summary"])
	}
}

func installTestWriter() *captureWriter {
	writer := &captureWriter{}
	logx.Reset()
	logx.SetWriter(writer)
	logx.SetLevel(logx.InfoLevel)
	return writer
}
