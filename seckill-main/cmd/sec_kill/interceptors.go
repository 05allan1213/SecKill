package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/BitofferHub/pkg/constant"
	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	bitlog "github.com/BitofferHub/seckill/internal/log"
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
		request, userID := summarizeSeckillRequest(req)
		if userID != 0 {
			ctx = bitlog.WithUser(ctx, userID)
		}

		begin := time.Now()
		reply, err := handler(ctx, req)
		bitlog.Access(ctx, detail, bitlog.AccessEntry{
			Action:   action,
			Method:   info.FullMethod,
			Code:     extractSeckillCode(reply),
			Request:  request,
			Response: summarizeSeckillResponse(reply),
			Err:      err,
			Cost:     time.Since(begin),
		})
		return reply, err
	}
}

func summarizeSeckillRequest(req any) (map[string]any, int64) {
	switch v := req.(type) {
	case *pb.SecKillV1Request:
		return map[string]any{"userID": v.UserID, "goodsNum": v.GoodsNum, "num": v.Num}, v.UserID
	case *pb.SecKillV2Request:
		return map[string]any{"userID": v.UserID, "goodsNum": v.GoodsNum, "num": v.Num}, v.UserID
	case *pb.SecKillV3Request:
		return map[string]any{"userID": v.UserID, "goodsNum": v.GoodsNum, "num": v.Num}, v.UserID
	case *pb.GetSecKillInfoRequest:
		return map[string]any{"userID": v.UserID, "secNum": v.SecNum}, v.UserID
	case *pb.GetGoodsListRequest:
		return map[string]any{"userID": v.UserID, "offset": v.Offset, "limit": v.Limit}, v.UserID
	default:
		return nil, 0
	}
}

func summarizeSeckillResponse(reply any) any {
	switch v := reply.(type) {
	case *pb.SecKillV1Reply:
		if v == nil {
			return nil
		}
		orderNum := ""
		if v.Data != nil {
			orderNum = v.Data.OrderNum
		}
		return map[string]any{"code": v.Code, "message": v.Message, "orderNum": orderNum}
	case *pb.SecKillV2Reply:
		if v == nil {
			return nil
		}
		orderNum := ""
		if v.Data != nil {
			orderNum = v.Data.OrderNum
		}
		return map[string]any{"code": v.Code, "message": v.Message, "orderNum": orderNum}
	case *pb.SecKillV3Reply:
		if v == nil {
			return nil
		}
		secNum := ""
		if v.Data != nil {
			secNum = v.Data.SecNum
		}
		return map[string]any{"code": v.Code, "message": v.Message, "secNum": secNum}
	case *pb.GetSecKillInfoReply:
		if v == nil || v.Data == nil {
			return nil
		}
		return map[string]any{"code": v.Code, "message": v.Message, "status": v.Data.Status, "orderNum": v.Data.OrderNum, "secNum": v.Data.SecNum, "reason": v.Data.Reason}
	case *pb.GetGoodsListReply:
		if v == nil || v.Data == nil {
			return nil
		}
		return map[string]any{"code": v.Code, "message": v.Message, "goodsCount": len(v.Data.GoodsList)}
	default:
		return nil
	}
}

func extractSeckillCode(reply any) int {
	switch v := reply.(type) {
	case *pb.SecKillV1Reply:
		if v != nil {
			return int(v.Code)
		}
	case *pb.SecKillV2Reply:
		if v != nil {
			return int(v.Code)
		}
	case *pb.SecKillV3Reply:
		if v != nil {
			return int(v.Code)
		}
	case *pb.GetGoodsListReply:
		if v != nil {
			return int(v.Code)
		}
	case *pb.GetSecKillInfoReply:
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
