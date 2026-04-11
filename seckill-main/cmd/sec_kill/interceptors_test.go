package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/BitofferHub/pkg/constant"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
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

func TestNewAccessLogInterceptor_Success(t *testing.T) {
	writer := installTestWriter()

	ctx := context.WithValue(context.Background(), constant.TraceID, "trace-success")
	req := struct{ GoodsNum string }{GoodsNum: "GOODS001"}
	reply := struct{ OrderNum string }{OrderNum: "ORDER001"}

	interceptor := NewAccessLogInterceptor(128)
	info := &grpc.UnaryServerInfo{FullMethod: "/sec_kill.SecKill/SecKillV1"}
	_, err := interceptor(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return reply, nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(writer.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(writer.entries))
	}

	entry := writer.entries[0]
	if entry.level != "info" {
		t.Fatalf("expected info level, got %s", entry.level)
	}
	if entry.msg != "rpc call completed" {
		t.Fatalf("unexpected message: %#v", entry.msg)
	}
	if entry.fields["method"] != info.FullMethod {
		t.Fatalf("unexpected method field: %#v", entry.fields["method"])
	}
	if entry.fields["grpc_code"] != "OK" {
		t.Fatalf("unexpected grpc_code: %#v", entry.fields["grpc_code"])
	}
	if entry.fields[constant.TraceID] != "trace-success" {
		t.Fatalf("unexpected trace field: %#v", entry.fields[constant.TraceID])
	}
	if summary, _ := entry.fields["req_summary"].(string); summary == "" {
		t.Fatalf("req_summary should not be empty")
	}
	if summary, _ := entry.fields["reply_summary"].(string); summary == "" {
		t.Fatalf("reply_summary should not be empty")
	}
}

func TestNewAccessLogInterceptor_Error(t *testing.T) {
	writer := installTestWriter()

	ctx := context.WithValue(context.Background(), constant.TraceID, "trace-error")
	req := struct{ GoodsNum string }{GoodsNum: "GOODS001"}
	expectedErr := errors.New("boom")

	interceptor := NewAccessLogInterceptor(128)
	info := &grpc.UnaryServerInfo{FullMethod: "/sec_kill.SecKill/SecKillV1"}
	_, err := interceptor(ctx, req, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(writer.entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(writer.entries))
	}

	entry := writer.entries[0]
	if entry.level != "error" {
		t.Fatalf("expected error level, got %s", entry.level)
	}
	if entry.msg != "rpc call failed" {
		t.Fatalf("unexpected message: %#v", entry.msg)
	}
	if entry.fields["grpc_code"] != "Unknown" {
		t.Fatalf("unexpected grpc_code: %#v", entry.fields["grpc_code"])
	}
	if gotErr, ok := entry.fields["err"].(error); !ok || !errors.Is(gotErr, expectedErr) {
		t.Fatalf("unexpected err field: %#v", entry.fields["err"])
	}
}

func installTestWriter() *captureWriter {
	writer := &captureWriter{}
	logx.Reset()
	logx.SetWriter(writer)
	logx.SetLevel(logx.InfoLevel)
	return writer
}

func TestNewAccessLogInterceptor_TruncatesSummaries(t *testing.T) {
	writer := installTestWriter()

	req := struct{ Payload string }{Payload: strings.Repeat("a", 64)}
	reply := struct{ Payload string }{Payload: strings.Repeat("b", 64)}

	interceptor := NewAccessLogInterceptor(32)
	_, err := interceptor(context.Background(), req, &grpc.UnaryServerInfo{FullMethod: "/sec_kill.SecKill/Test"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return reply, nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	entry := writer.entries[0]
	reqSummary := entry.fields["req_summary"].(string)
	replySummary := entry.fields["reply_summary"].(string)
	if len(reqSummary) > 32 || len(replySummary) > 32 {
		t.Fatalf("expected summaries to be truncated, got req=%q reply=%q", reqSummary, replySummary)
	}
	if !strings.HasSuffix(reqSummary, "...(truncated)") || !strings.HasSuffix(replySummary, "...(truncated)") {
		t.Fatalf("expected truncation suffix, got req=%q reply=%q", reqSummary, replySummary)
	}
}
