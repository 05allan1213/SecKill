package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	pb "github.com/BitofferHub/seckill/api/sec_kill/proto"
	"github.com/BitofferHub/seckill/internal/config"
	healthserver "github.com/BitofferHub/seckill/internal/health"
	seckillserver "github.com/BitofferHub/seckill/internal/server/seckill"
	"github.com/BitofferHub/seckill/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "go.uber.org/automaxprocs"
)

var configFile = flag.String("f", "etc/seckill.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	config.ApplyEnvOverrides(&c)

	svcCtx := svc.NewServiceContext(c)
	defer svcCtx.Data.Close()

	consumerRunner := NewConsumerRunner(svcCtx)
	if err := consumerRunner.Start(); err != nil {
		panic(err)
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = consumerRunner.Stop(stopCtx)
	}()

	healthSrv := healthserver.NewServer(c.HealthProbe.Host, c.HealthProbe.Port, c.Name, []healthserver.Check{
		{
			Name: "mysql",
			Run: func(ctx context.Context) error {
				return svcCtx.Data.PingDB(ctx)
			},
		},
		{
			Name: "redis",
			Run: func(ctx context.Context) error {
				return svcCtx.Data.PingRedis(ctx)
			},
		},
		{
			Name: "kafka",
			Run: func(ctx context.Context) error {
				return svcCtx.Data.PingKafkaProducer(ctx)
			},
		},
		{
			Name: "consumer",
			Run: func(context.Context) error {
				if !consumerRunner.Running() {
					return fmt.Errorf("consumer not running")
				}
				return nil
			},
		},
	})
	healthSrv.Start()
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = healthSrv.Shutdown(stopCtx)
	}()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterSecKillServer(grpcServer, seckillserver.NewSecKillServer(svcCtx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(NewTraceIDInterceptor(), NewAccessLogInterceptor(c.Log.AccessDetail))
	defer s.Stop()

	fmt.Printf("Starting seckill rpc server at %s...\n", c.ListenOn)
	s.Start()
}
