package main

import (
	"flag"
	"fmt"

	"github.com/BitofferHub/gateway/internal/config"
	"github.com/BitofferHub/gateway/internal/handler"
	gwmiddleware "github.com/BitofferHub/gateway/internal/middleware"
	"github.com/BitofferHub/gateway/internal/svc"
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

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	server.Use(gwmiddleware.NewTraceMiddleware().Handle)

	svcCtx := svc.NewServiceContext(c)
	defer svcCtx.Close()
	handler.RegisterHandlers(server, svcCtx)

	fmt.Printf("Starting gateway %s at %s:%d...\n", Version, c.Host, c.Port)
	server.Start()
}
