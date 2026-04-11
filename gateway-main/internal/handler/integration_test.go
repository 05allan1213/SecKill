//go:build integration

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BitofferHub/gateway/internal/config"
	gwmiddleware "github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"
	secproto "github.com/BitofferHub/seckill/api/sec_kill/proto"
	userv1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func TestGatewayHandlersIntegration(t *testing.T) {
	userServer := grpc.NewServer()
	userv1.RegisterUserServer(userServer, fakeUserServer{})
	userListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen user rpc failed: %v", err)
	}
	defer userListener.Close()
	go userServer.Serve(userListener)
	defer userServer.Stop()

	seckillServer := grpc.NewServer()
	secproto.RegisterSecKillServer(seckillServer, fakeSeckillServer{})
	seckillListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen seckill rpc failed: %v", err)
	}
	defer seckillListener.Close()
	go seckillServer.Serve(seckillListener)
	defer seckillServer.Stop()

	miniRedis, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis failed: %v", err)
	}
	defer miniRedis.Close()

	cfg := config.Config{
		Auth:  config.AuthConf{Secret: "secret key", Timeout: time.Hour},
		Redis: config.RedisConf{Addr: miniRedis.Addr()},
		UserRpc: zrpc.RpcClientConf{
			Target:  userListener.Addr().String(),
			Timeout: 2000,
		},
		SeckillRpc: zrpc.RpcClientConf{
			Target:  seckillListener.Addr().String(),
			Timeout: 2000,
		},
		RoutePolicies: map[string]config.RoutePolicy{
			"/bitstorm/get_user_info_by_name": {LimitRate: 1000, LimitTimeout: 2000, RetryTime: 50},
			"/bitstorm/v1/sec_kill":           {LimitRate: 1000, LimitTimeout: 2000, RetryTime: 50},
		},
	}

	svcCtx := svc.NewServiceContext(cfg)
	defer svcCtx.Close()

	auth := gwmiddleware.NewAuthMiddleware(svcCtx.AuthConfig)
	userLimit := gwmiddleware.NewRouteLimitMiddleware(svcCtx, "/bitstorm/get_user_info_by_name")
	seckillLimit := gwmiddleware.NewRouteLimitMiddleware(svcCtx, "/bitstorm/v1/sec_kill")

	mux := http.NewServeMux()
	mux.HandleFunc("/login", LoginHandler(svcCtx))
	mux.HandleFunc("/bitstorm/get_user_info_by_name", auth.Handle(userLimit.Handle(BitstormGetUserInfoByNameHandler(svcCtx))))
	mux.HandleFunc("/bitstorm/v1/sec_kill", auth.Handle(seckillLimit.Handle(BitstormSecKillV1Handler(svcCtx))))

	server := httptest.NewServer(gwmiddleware.NewTraceMiddleware().Handle(mux.ServeHTTP))
	defer server.Close()

	loginResp, err := http.Post(server.URL+"/login", "application/json", bytes.NewBufferString(`{"username":"admin","password":"123321"}`))
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer loginResp.Body.Close()

	var loginReply map[string]any
	if err := json.NewDecoder(loginResp.Body).Decode(&loginReply); err != nil {
		t.Fatalf("decode login reply failed: %v", err)
	}
	token, _ := loginReply["token"].(string)
	if token == "" {
		t.Fatalf("unexpected login reply: %#v", loginReply)
	}

	userReq, _ := http.NewRequest(http.MethodGet, server.URL+"/bitstorm/get_user_info_by_name?user_name=admin", nil)
	userReq.Header.Set("Authorization", "Bearer "+token)
	userReq.Header.Set("Trace-ID", "itest-user")
	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		t.Fatalf("get user request failed: %v", err)
	}
	defer userResp.Body.Close()
	if userResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected user response status: %d", userResp.StatusCode)
	}

	secReq, _ := http.NewRequest(http.MethodPost, server.URL+"/bitstorm/v1/sec_kill", bytes.NewBufferString(`{"goodsNum":"abc123","num":1}`))
	secReq.Header.Set("Authorization", "Bearer "+token)
	secReq.Header.Set("Content-Type", "application/json")
	secReq.Header.Set("Trace-ID", "itest-sec")
	secResp, err := http.DefaultClient.Do(secReq)
	if err != nil {
		t.Fatalf("seckill request failed: %v", err)
	}
	defer secResp.Body.Close()
	if secResp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected seckill response status: %d", secResp.StatusCode)
	}
}

type fakeUserServer struct {
	userv1.UnimplementedUserServer
}

func (fakeUserServer) GetUser(context.Context, *userv1.GetUserRequest) (*userv1.GetUserReply, error) {
	return &userv1.GetUserReply{
		Code: 200,
		Data: &userv1.GetUserReplyData{
			UserID:   "1",
			UserName: "admin",
			Pwd:      "123321",
		},
	}, nil
}

func (fakeUserServer) GetUserByName(context.Context, *userv1.GetUserByNameRequest) (*userv1.GetUserByNameReply, error) {
	return &userv1.GetUserByNameReply{
		Code: 200,
		Data: &userv1.GetUserReplyData{
			UserID:   "1",
			UserName: "admin",
			Pwd:      "123321",
		},
	}, nil
}

type fakeSeckillServer struct {
	secproto.UnimplementedSecKillServer
}

func (fakeSeckillServer) SecKillV1(context.Context, *secproto.SecKillV1Request) (*secproto.SecKillV1Reply, error) {
	return &secproto.SecKillV1Reply{
		Code: 200,
		Data: &secproto.SecKillV1ReplyData{OrderNum: "itest-order"},
	}, nil
}
