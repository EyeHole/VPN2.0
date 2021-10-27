package logs

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"VPN2.0/server/internal/config"
)

func BuildLogger(conf *config.Config) *zap.Logger {
	var zapCfg = zap.NewProductionConfig()
	zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if conf.Debug {
		zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}

	return logger
}
