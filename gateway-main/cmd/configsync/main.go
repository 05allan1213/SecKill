package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	endpoints = flag.String("endpoints", "127.0.0.1:20001", "comma separated etcd endpoints")
	sourceDir = flag.String("dir", "../configs/runtime", "runtime config directory")
)

var keyByFile = map[string]string{
	"gateway.yaml": "/bitstorm/gateway/runtime",
	"gateway.yml":  "/bitstorm/gateway/runtime",
	"user.yaml":    "/bitstorm/user/runtime",
	"user.yml":     "/bitstorm/user/runtime",
	"seckill.yaml": "/bitstorm/seckill/runtime",
	"seckill.yml":  "/bitstorm/seckill/runtime",
}

func main() {
	flag.Parse()

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   splitCSV(*endpoints),
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	entries, err := os.ReadDir(*sourceDir)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		key, ok := keyByFile[entry.Name()]
		if !ok {
			continue
		}

		path := filepath.Join(*sourceDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			panic(err)
		}
		if _, err := client.Put(ctx, key, string(content)); err != nil {
			panic(err)
		}
		fmt.Printf("synced %s -> %s\n", path, key)
	}
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
