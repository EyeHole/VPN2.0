package main

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"VPN2.0/client/internal/client"
	"VPN2.0/client/internal/config"
	"VPN2.0/client/internal/logs"
)

func main() {
	conf, err := config.New()
	if err != nil {
		panic(err)
	}

	logger := logs.BuildLogger(conf)
	ctx := ctxzap.ToContext(context.Background(), logger)
	logger.Info("server starting...")

	err = client.RunClient(ctx)
	if err != nil {
		panic(err)
	}
}
