package lib

//import (
//	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
//	uuid "github.com/nu7hatch/gouuid"
//	"go.uber.org/zap"
//	"go.uber.org/zap/zapcore"
//	"golang.org/x/net/context"
//	"google.golang.org/grpc/metadata"
//)
//
//type ctxvalue string
//
//const (
//	KeyRequestID ctxvalue = "request_id"
//)
//
//func buildContext(ctx context.Context) *zap.Logger {
//	ctxNew := deriveRequestContext(*ctx)
//	(*ctx) = ctxNew
//
//	logger, err := zapCfg.Build()
//	if err != nil {
//		panic(err)
//	}
//
//	return logger
//}
//
//func deriveRequestContext(ctx context.Context) context.Context {
//	var (
//		requestID string
//	)
//	// try getting request IDs from grpc metadata
//	if md, ok := metadata.FromIncomingContext(ctx); ok {
//		if v := md.Get(string(KeyRequestID)); len(v) > 0 {
//			requestID = v[0]
//		}
//	}
//
//	// generate it ourself if it's not set
//	if requestID == "" {
//		requestID = "svc-" + uuid.NewV4().String()
//	}
//
//	ctx = context.WithValue(ctx, KeyRequestID, requestID)
//
//	// also add them to outgoing context, for any further calls to services
//	ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
//		string(KeyRequestID): requestID,
//	}))
//
//	// add request and organization IDs to zap logger which is already in context
//	logFields := []zapcore.Field{
//		zap.String(string(KeyRequestID), requestID),
//	}
//	AddLoggerFields(ctx, logFields...)
//
//	// return resulting context
//	return ctx
//}
//
//// AddLoggerFields is a proxy for ctxzap.AddFields
//func AddLoggerFields(ctx context.Context, fields ...zapcore.Field) {
//	ctxzap.AddFields(ctx, fields...)
//}
//
//// GetRequestID extracts request-scoped request id from context
//func GetRequestID(ctx context.Context) string {
//	if v := ctx.Value(KeyRequestID); v != nil {
//		return v.(string)
//	}
//	return "n/a"
//}
