package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	legacydiscovery "github.com/BitofferHub/pkg/middlewares/discovery"
	"github.com/BitofferHub/seckill/internal/config"
	"github.com/go-kratos/kratos/v2/registry"
)

var compatServiceNames = []string{"sec_kill-svr"}

func RegisterCompatServices(c config.Config) (func(), error) {
	registrar := legacydiscovery.NewRegistrar(c.Etcd.Hosts)
	host := normalizeAdvertiseHost(c.ListenOn)
	grpcEndpoint := fmt.Sprintf("grpc://%s", net.JoinHostPort(host, portOnly(c.ListenOn)))
	httpEndpoint := fmt.Sprintf("http://%s", net.JoinHostPort(host, portOnly(c.CompatHttp.Addr)))

	instances := make([]*registry.ServiceInstance, 0, len(compatServiceNames))
	ctx := context.Background()

	for _, name := range compatServiceNames {
		instance := &registry.ServiceInstance{
			ID:       fmt.Sprintf("%s-%s", name, hostID()),
			Name:     name,
			Version:  "gozero-phase2",
			Metadata: map[string]string{"migration": "phase2-seckill-main"},
			Endpoints: []string{
				grpcEndpoint,
				httpEndpoint,
			},
		}
		if err := registrar.Register(ctx, instance); err != nil {
			return nil, err
		}
		instances = append(instances, instance)
	}

	return func() {
		for _, instance := range instances {
			_ = registrar.Deregister(ctx, instance)
		}
	}, nil
}

func normalizeAdvertiseHost(listenOn string) string {
	host, _, err := net.SplitHostPort(listenOn)
	if err != nil {
		return "127.0.0.1"
	}
	switch host {
	case "", "0.0.0.0", "::":
		return firstNonLoopbackIPv4()
	default:
		return host
	}
}

func portOnly(listenOn string) string {
	_, port, err := net.SplitHostPort(listenOn)
	if err != nil {
		return ""
	}
	return port
}

func hostID() string {
	id, err := os.Hostname()
	if err != nil || id == "" {
		return "seckill-main"
	}
	return id
}

func firstNonLoopbackIPv4() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			ip = ip.To4()
			if ip == nil || ip.IsLoopback() {
				continue
			}
			return strings.TrimSpace(ip.String())
		}
	}
	return "127.0.0.1"
}
