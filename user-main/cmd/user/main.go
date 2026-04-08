package main

import (
	"context"
	"flag"
	"fmt"

	userserver "github.com/BitofferHub/user/internal/server/user"
	"github.com/BitofferHub/user/internal/config"
	"github.com/BitofferHub/user/internal/svc"
	v1 "github.com/BitofferHub/user/api/user/v1"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	svcCtx := svc.NewServiceContext(c)

	compatHTTP, err := StartCompatHTTPServer(c, svcCtx)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = compatHTTP.Shutdown(context.Background())
	}()

	cleanupCompat, err := RegisterCompatServices(c)
	if err != nil {
		panic(err)
	}
	defer cleanupCompat()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		v1.RegisterUserServer(grpcServer, userserver.NewUserServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(NewTraceIDInterceptor(), NewAccessLogInterceptor())
	defer s.Stop()

	fmt.Printf("Starting user rpc server at %s...\n", c.ListenOn)
	logx.Infof("compatibility http server listening on %s", c.CompatHttp.Addr)
	s.Start()
}
