package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BitofferHub/pkg/constant"
	v1 "github.com/BitofferHub/user/api/user/v1"
	bitlog "github.com/BitofferHub/user/internal/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func NewTraceIDInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		traceID := extractTraceID(ctx)
		if traceID != "" {
			ctx = context.WithValue(ctx, constant.TraceID, traceID)
			ctx = bitlog.WithTrace(ctx, traceID)
		}
		return handler(ctx, req)
	}
}

func NewAccessLogInterceptor(detail string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		action := info.FullMethod
		ctx = bitlog.WithAction(ctx, action)
		request, userID := summarizeUserRequest(req)
		if userID != "" {
			ctx = bitlog.WithUser(ctx, userID)
		}

		begin := time.Now()
		reply, err := handler(ctx, req)
		bitlog.Access(ctx, detail, bitlog.AccessEntry{
			Action:   action,
			Method:   info.FullMethod,
			Code:     extractUserCode(reply),
			Request:  request,
			Response: summarizeUserResponse(reply),
			Err:      err,
			Cost:     time.Since(begin),
		})
		return reply, err
	}
}

func summarizeUserRequest(req any) (map[string]any, string) {
	switch v := req.(type) {
	case *v1.CreateUserRequest:
		return map[string]any{"userName": v.UserName}, ""
	case *v1.GetUserRequest:
		return map[string]any{"userID": v.UserID}, fmt.Sprintf("%d", v.UserID)
	case *v1.GetUserByNameRequest:
		return map[string]any{"userName": v.UserName}, ""
	default:
		return nil, ""
	}
}

func summarizeUserResponse(reply any) any {
	switch v := reply.(type) {
	case *v1.CreateUserReply:
		if v == nil {
			return nil
		}
		return map[string]any{"code": v.Code, "message": v.Message}
	case *v1.GetUserReply:
		if v == nil {
			return nil
		}
		return map[string]any{"code": v.Code, "message": v.Message}
	case *v1.GetUserByNameReply:
		if v == nil {
			return nil
		}
		return map[string]any{"code": v.Code, "message": v.Message}
	default:
		return nil
	}
}

func extractUserCode(reply any) int {
	switch v := reply.(type) {
	case *v1.CreateUserReply:
		if v != nil {
			return int(v.Code)
		}
	case *v1.GetUserReply:
		if v != nil {
			return int(v.Code)
		}
	case *v1.GetUserByNameReply:
		if v != nil {
			return int(v.Code)
		}
	}
	return 0
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
