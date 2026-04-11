package main

import (
	"flag"
	"fmt"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/config"
	userserver "github.com/BitofferHub/user/internal/server/user"
	"github.com/BitofferHub/user/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
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
	config.ApplyEnvOverrides(&c)

	svcCtx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		v1.RegisterUserServer(grpcServer, userserver.NewUserServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(NewTraceIDInterceptor(), NewAccessLogInterceptor(c.Log.AccessDetail))
	defer s.Stop()

	fmt.Printf("Starting user rpc server at %s...\n", c.ListenOn)
	s.Start()
}
