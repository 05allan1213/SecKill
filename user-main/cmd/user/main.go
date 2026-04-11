package main

import (
	"flag"
	"fmt"

	v1 "github.com/BitofferHub/user/api/user/v1"
	"github.com/BitofferHub/user/internal/config"
	userserver "github.com/BitofferHub/user/internal/server/user"
	"github.com/BitofferHub/user/internal/svc"

	obs "github.com/BitofferHub/observability"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/user.yaml", "the config file")

func main() {
	flag.Parse()

	c, err := config.Load(*configFile)
	if err != nil {
		panic(err)
	}

	logWriter, err := obs.NewWriter(c.Log.Path, obs.RotationConfig{
		MaxSizeMB:  c.Observability.LogRotation.MaxSizeMB,
		MaxBackups: c.Observability.LogRotation.MaxBackups,
		KeepDays:   c.Log.KeepDays,
		Compress:   c.Observability.LogRotation.Compress,
	})
	if err != nil {
		panic(err)
	}
	defer logWriter.Close()
	logx.SetWriter(logWriter)

	svcCtx := svc.NewServiceContext(c)
	defer svcCtx.Close()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		v1.RegisterUserServer(grpcServer, userserver.NewUserServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	logx.SetWriter(logWriter)
	s.AddUnaryInterceptors(NewTraceIDInterceptor(), NewAccessLogInterceptor(c.Observability.AccessLog.SummaryMaxBytes))
	defer s.Stop()

	fmt.Printf("Starting user rpc server at %s...\n", c.ListenOn)
	s.Start()
}
