package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	obs "github.com/BitofferHub/observability"
	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/config"
	seckillserver "github.com/BitofferHub/seckill/internal/server/seckill"
	"github.com/BitofferHub/seckill/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "go.uber.org/automaxprocs"
)

var configFile = flag.String("f", "etc/seckill.yaml", "the config file")

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

	consumerRunner := NewConsumerRunner(svcCtx)
	if err := consumerRunner.Start(); err != nil {
		panic(err)
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = consumerRunner.Stop(stopCtx)
	}()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterSecKillServer(grpcServer, seckillserver.NewSecKillServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	logx.SetWriter(logWriter)
	s.AddUnaryInterceptors(NewTraceIDInterceptor(), NewAccessLogInterceptor(c.Observability.AccessLog.SummaryMaxBytes))
	defer s.Stop()

	fmt.Printf("Starting seckill rpc server at %s...\n", c.ListenOn)
	s.Start()
}
