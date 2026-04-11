package log

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/BitofferHub/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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

func TestGormLoggerTrace_Query(t *testing.T) {
	writer := installLogTestWriter()
	logger := NewGormLogger(1000)
	ctx := context.WithValue(context.Background(), constant.TraceID, "trace-query")

	logger.Trace(ctx, time.Now().Add(-5*time.Millisecond), func() (string, int64) {
		return "select   * from t_user_info", 1
	}, nil)

	if len(writer.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(writer.entries))
	}
	entry := writer.entries[0]
	if entry.level != "info" {
		t.Fatalf("expected info level, got %s", entry.level)
	}
	if entry.msg != "gorm query completed" {
		t.Fatalf("unexpected message: %#v", entry.msg)
	}
	if entry.fields["sql"] != "select * from t_user_info" {
		t.Fatalf("unexpected sql field: %#v", entry.fields["sql"])
	}
	if entry.fields[constant.TraceID] != "trace-query" {
		t.Fatalf("unexpected trace field: %#v", entry.fields[constant.TraceID])
	}
}

func TestGormLoggerTrace_SlowQuery(t *testing.T) {
	writer := installLogTestWriter()
	logger := NewGormLogger(1)

	logger.Trace(context.Background(), time.Now().Add(-5*time.Millisecond), func() (string, int64) {
		return "select * from t_user_info", 3
	}, nil)

	if len(writer.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(writer.entries))
	}
	entry := writer.entries[0]
	if entry.level != "slow" {
		t.Fatalf("expected slow level, got %s", entry.level)
	}
	if entry.msg != "gorm slow query" {
		t.Fatalf("unexpected message: %#v", entry.msg)
	}
}

func TestGormLoggerTrace_Error(t *testing.T) {
	writer := installLogTestWriter()
	logger := NewGormLogger(1000)
	expectedErr := errors.New("db failed")

	logger.Trace(context.Background(), time.Now().Add(-1*time.Millisecond), func() (string, int64) {
		return "select * from t_user_info", 0
	}, expectedErr)

	if len(writer.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(writer.entries))
	}
	entry := writer.entries[0]
	if entry.level != "error" {
		t.Fatalf("expected error level, got %s", entry.level)
	}
	if entry.msg != "gorm query failed" {
		t.Fatalf("unexpected message: %#v", entry.msg)
	}
	if gotErr, ok := entry.fields["err"].(error); !ok || !errors.Is(gotErr, expectedErr) {
		t.Fatalf("unexpected err field: %#v", entry.fields["err"])
	}
}

func TestGormLoggerTrace_RecordNotFound(t *testing.T) {
	writer := installLogTestWriter()
	logger := NewGormLogger(1000)

	logger.Trace(context.Background(), time.Now().Add(-1*time.Millisecond), func() (string, int64) {
		return "select * from t_user_info where id = 1", 0
	}, gorm.ErrRecordNotFound)

	if len(writer.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(writer.entries))
	}
	entry := writer.entries[0]
	if entry.level != "info" {
		t.Fatalf("expected info level, got %s", entry.level)
	}
	if entry.msg != "gorm record not found" {
		t.Fatalf("unexpected message: %#v", entry.msg)
	}
}

func installLogTestWriter() *captureWriter {
	writer := &captureWriter{}
	logx.Reset()
	logx.SetWriter(writer)
	logx.SetLevel(logx.InfoLevel)
	return writer
}
