package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BitofferHub/pkg/constant"
	bitlog "github.com/BitofferHub/user/internal/log"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewTraceIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		traceID := extractTraceID(ctx)
		if traceID != "" {
			ctx = context.WithValue(ctx, constant.TraceID, traceID)
		}
		return handler(ctx, req)
	}
}

func NewAccessLogInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		begin := time.Now()
		reply, err := handler(ctx, req)
		fields := []logx.LogField{
			logx.Field("method", info.FullMethod),
			logx.Field("duration_ms", time.Since(begin).Milliseconds()),
			logx.Field("grpc_code", status.Code(err).String()),
			logx.Field("req_summary", summarizePayload(req)),
			logx.Field("reply_summary", summarizePayload(reply)),
		}
		if err != nil {
			fields = append(fields, logx.Field("err", err))
			bitlog.ErrorContextw(ctx, "rpc call failed", fields...)
		} else {
			bitlog.InfoContextw(ctx, "rpc call completed", fields...)
		}
		return reply, err
	}
}

func extractTraceID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	keys := []string{
		strings.ToLower(constant.TraceID),
		fmt.Sprintf("x-md-global-%s", strings.ToLower(constant.TraceID)),
		"x-md-global-traceid",
		"traceid",
		"trace-id",
	}
	for _, key := range keys {
		if values := md.Get(key); len(values) > 0 && values[0] != "" {
			return values[0]
		}
	}
	return ""
}

func summarizePayload(v interface{}) string {
	if v == nil {
		return "<nil>"
	}

	summary := fmt.Sprintf("%T %v", v, v)
	const maxLen = 512
	if len(summary) <= maxLen {
		return summary
	}

	return summary[:maxLen] + "...(truncated)"
}
