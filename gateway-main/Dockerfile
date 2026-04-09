FROM golang:1.26.1 AS builder

COPY . /src
WORKDIR /src

RUN GOPROXY=https://goproxy.cn go mod tidy && \
	GOPROXY=https://goproxy.cn make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app

WORKDIR /app

EXPOSE 8998
VOLUME /data/conf

CMD ["./gateway", "-f", "/data/conf/gateway.yaml"]
