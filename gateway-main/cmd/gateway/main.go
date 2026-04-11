package main

import (
	"flag"
	"fmt"

	"github.com/BitofferHub/gateway/internal/config"
	"github.com/BitofferHub/gateway/internal/handler"
	gwmiddleware "github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"
	obs "github.com/BitofferHub/observability"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	_ "go.uber.org/automaxprocs"
)

var (
	Version    string
	configFile = flag.String("f", "etc/gateway.yaml", "the config file")
)

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

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()
	logx.SetWriter(logWriter)

	server.Use(gwmiddleware.NewTraceMiddleware().Handle)

	svcCtx := svc.NewServiceContext(c)
	defer svcCtx.Close()
	handler.RegisterHandlers(server, svcCtx)

	fmt.Printf("Starting gateway %s at %s:%d...\n", Version, c.Host, c.Port)
	server.Start()
}
