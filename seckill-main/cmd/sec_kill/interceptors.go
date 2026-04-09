package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BitofferHub/pkg/constant"
	bitlog "github.com/BitofferHub/pkg/middlewares/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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
		traceID, _ := ctx.Value(constant.TraceID).(string)
		bitlog.InfoContextf(ctx, "traceID:%s method:%s req:%+v cost:%v err:%v reply:%+v",
			traceID, info.FullMethod, req, time.Since(begin), err, reply)
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
