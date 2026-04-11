package log

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/BitofferHub/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
)

type gormLogEntry struct {
	level  string
	msg    interface{}
	fields map[string]interface{}
}

type gormCaptureWriter struct {
	entries []gormLogEntry
}

func (w *gormCaptureWriter) Alert(v any)                         {}
func (w *gormCaptureWriter) Severe(v any)                        {}
func (w *gormCaptureWriter) Stack(v any)                         {}
func (w *gormCaptureWriter) Stat(v any, fields ...logx.LogField) {}
func (w *gormCaptureWriter) Close() error                        { return nil }

func (w *gormCaptureWriter) Debug(v any, fields ...logx.LogField) {
	w.entries = append(w.entries, newGormLogEntry("debug", v, fields...))
}

func (w *gormCaptureWriter) Error(v any, fields ...logx.LogField) {
	w.entries = append(w.entries, newGormLogEntry("error", v, fields...))
}

func (w *gormCaptureWriter) Info(v any, fields ...logx.LogField) {
	w.entries = append(w.entries, newGormLogEntry("info", v, fields...))
}

func (w *gormCaptureWriter) Slow(v any, fields ...logx.LogField) {
	w.entries = append(w.entries, newGormLogEntry("slow", v, fields...))
}

func newGormLogEntry(level string, msg any, fields ...logx.LogField) gormLogEntry {
	entry := gormLogEntry{
		level:  level,
		msg:    msg,
		fields: make(map[string]interface{}, len(fields)),
	}
	for _, field := range fields {
		entry.fields[field.Key] = field.Value
	}
	return entry
}

func TestGormLoggerTrace_CompletedQuery(t *testing.T) {
	writer := installGormTestWriter()

	logger := NewGormLogger(50)
	ctx := context.WithValue(context.Background(), constant.TraceID, "trace-query")
	logger.Trace(ctx, time.Now().Add(-10*time.Millisecond), func() (string, int64) {
		return "SELECT *   FROM t_goods WHERE goods_num = ?", 1
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
	if entry.fields[constant.TraceID] != "trace-query" {
		t.Fatalf("unexpected trace field: %#v", entry.fields[constant.TraceID])
	}
	if entry.fields["sql"] != "SELECT * FROM t_goods WHERE goods_num = ?" {
		t.Fatalf("unexpected sql field: %#v", entry.fields["sql"])
	}
}

func TestGormLoggerTrace_SlowQuery(t *testing.T) {
	writer := installGormTestWriter()

	logger := NewGormLogger(1)
	logger.Trace(context.Background(), time.Now().Add(-5*time.Millisecond), func() (string, int64) {
		return "SELECT * FROM t_goods", 2
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

func TestGormLoggerTrace_ErrorQuery(t *testing.T) {
	writer := installGormTestWriter()

	logger := NewGormLogger(50)
	expectedErr := errors.New("db down")
	logger.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT * FROM t_goods", 0
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

func installGormTestWriter() *gormCaptureWriter {
	writer := &gormCaptureWriter{}
	logx.Reset()
	logx.SetWriter(writer)
	logx.SetLevel(logx.InfoLevel)
	return writer
}
