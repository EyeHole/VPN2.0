package ctxmeta

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

// GetLogger extracts request-scoped zap logger from context
func GetLogger(ctx context.Context) *zap.Logger {
	return ctxzap.Extract(ctx)
}
