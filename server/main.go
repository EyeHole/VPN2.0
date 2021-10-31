package main

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"

	"VPN2.0/server/internal/config"
	"VPN2.0/server/internal/logs"
	"VPN2.0/server/internal/server"
)

func main() {
	conf, err := config.New()
	if err != nil {
		panic(err)
	}

	logger := logs.BuildLogger(conf)
	ctx := ctxzap.ToContext(context.Background(), logger)
	logger.Info("server starting...")

	server.CreateServer(ctx)

	err = server.RunServer(ctx, conf.ServerAddr)
	if err != nil {
		panic(err)
	}
}
