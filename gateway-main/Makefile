VERSION=$(shell git describe --tags --always 2>/dev/null || git rev-parse --short HEAD)

.PHONY: init
init:
	go install github.com/zeromicro/go-zero/tools/goctl@latest

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/gateway ./cmd/gateway

.PHONY: fmt
fmt:
	gofmt -w ./cmd ./internal ./limiter

.PHONY: help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@echo ' init   install protoc/goctl tooling'
	@echo ' tidy   sync go modules'
	@echo ' build  build gateway binary'
	@echo ' fmt    format gateway source'

.DEFAULT_GOAL := help
