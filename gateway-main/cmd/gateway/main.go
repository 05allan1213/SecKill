package main

import (
	"flag"
	"fmt"

	"github.com/BitofferHub/gateway/internal/config"
	"github.com/BitofferHub/gateway/internal/handler"
	gwlog "github.com/BitofferHub/gateway/internal/log"
	gwmiddleware "github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	_ "go.uber.org/automaxprocs"
)

var (
	Version    string
	configFile = flag.String("f", "etc/gateway.yaml", "the config file")
)

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	gwlog.Info(nil, "gateway starting",
		gwlog.Field(gwlog.FieldAction, "gateway.start"),
		gwlog.Field("host", c.Host),
		gwlog.Field("port", c.Port),
	)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	server.Use(gwmiddleware.NewTraceMiddleware().Handle)

	svcCtx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, svcCtx)

	fmt.Printf("Starting gateway %s at %s:%d...\n", Version, c.Host, c.Port)
	server.Start()
}
