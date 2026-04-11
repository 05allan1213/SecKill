package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"
	secproto "github.com/BitofferHub/seckill/api/sec_kill/proto"
	userv1 "github.com/BitofferHub/user/api/user/v1"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type healthCheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type healthResponse struct {
	Status    string              `json:"status"`
	Service   string              `json:"service"`
	Checks    []healthCheckResult `json:"checks,omitempty"`
	Timestamp string              `json:"timestamp"`
}

type healthCheck struct {
	name string
	run  func(context.Context) error
}

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		service := "gateway"
		if svcCtx != nil && svcCtx.Config.Name != "" {
			service = svcCtx.Config.Name
		}
		middleware.WriteJSON(w, http.StatusOK, healthResponse{
			Status:    "ok",
			Service:   service,
			Timestamp: time.Now().Format(time.RFC3339),
		})
	}
}

func ReadyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	checks := []healthCheck{
		{name: "redis", run: func(ctx context.Context) error {
			if svcCtx == nil || svcCtx.Redis == nil {
				return errors.New("redis client not configured")
			}
			return svcCtx.Redis.Ping(ctx).Err()
		}},
		{name: "user_rpc", run: func(ctx context.Context) error {
			if svcCtx == nil || svcCtx.RPCClients == nil || svcCtx.UserClient == nil {
				return errors.New("user rpc client not configured")
			}
			_, err := svcCtx.UserClient.GetUser(ctx, &userv1.GetUserRequest{UserID: 1})
			return allowReadyRPCError(err, codes.NotFound)
		}},
		{name: "seckill_rpc", run: func(ctx context.Context) error {
			if svcCtx == nil || svcCtx.RPCClients == nil || svcCtx.SeckillClient == nil {
				return errors.New("seckill rpc client not configured")
			}
			_, err := svcCtx.SeckillClient.GetGoodsList(ctx, &secproto.GetGoodsListRequest{
				UserID: 1,
				Offset: 0,
				Limit:  1,
			})
			return allowReadyRPCError(err)
		}},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		service := "gateway"
		if svcCtx != nil && svcCtx.Config.Name != "" {
			service = svcCtx.Config.Name
		}
		resp := healthResponse{
			Status:    "ok",
			Service:   service,
			Timestamp: time.Now().Format(time.RFC3339),
			Checks:    make([]healthCheckResult, 0, len(checks)),
		}

		statusCode := http.StatusOK
		for _, check := range checks {
			checkCtx, cancel := context.WithTimeout(r.Context(), time.Second)
			err := check.run(checkCtx)
			cancel()

			result := healthCheckResult{Name: check.name, Status: "ok"}
			if err != nil {
				statusCode = http.StatusServiceUnavailable
				resp.Status = "not_ready"
				result.Status = "failed"
				result.Error = err.Error()
			}
			resp.Checks = append(resp.Checks, result)
		}

		middleware.WriteJSON(w, statusCode, resp)
	}
}

func allowReadyRPCError(err error, allowed ...codes.Code) error {
	if err == nil {
		return nil
	}
	st, ok := grpcstatus.FromError(err)
	if !ok {
		return err
	}
	for _, code := range allowed {
		if st.Code() == code {
			return nil
		}
	}
	return err
}
