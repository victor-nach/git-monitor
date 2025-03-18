package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/victor-nach/git-monitor/config"
)

const (
	serviceName = "git-monitor"
	version     = "1.0.0"
)

func New(appEnv string) (*zap.Logger, error) {
	var cfg zap.Config

	switch appEnv {
	case config.AppEnvDevelopment:
		cfg = zap.NewDevelopmentConfig()
		cfg.DisableStacktrace = true
		cfg.Encoding = "console" 
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder 

	default:
		cfg = zap.NewProductionConfig()
		cfg.DisableStacktrace = false
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

		cfg.Encoding = "console" 
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder 
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	logger = logger.With(
		zap.String("service", serviceName),
		zap.String("app_env", appEnv),
	)

	return logger, nil
}
