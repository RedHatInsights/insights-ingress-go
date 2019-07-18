package logger

import (
	"flag"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is an instance of the global zap.Logger
var Log *zap.Logger
var logLevel zapcore.Level

// InitLogger initializes the Entitlements API logger
func InitLogger() *zap.Logger {
	if Log == nil {
		viper.SetDefault("INGRESS_LOG_LEVEL", "INFO")
		viper.AutomaticEnv()
		switch viper.GetString("INGRESS_LOG_LEVEL") {
		case "DEBUG":
			logLevel = zapcore.DebugLevel
		case "ERROR":
			logLevel = zapcore.ErrorLevel
		default:
			logLevel = zapcore.InfoLevel
		}
		if flag.Lookup("test.v") != nil {
			logLevel = zapcore.FatalLevel
		}

		cfg := zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		logger, _ := zap.Config{
			Encoding:         "json",
			Level:            zap.NewAtomicLevelAt(logLevel),
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
			EncoderConfig:    cfg,
		}.Build()

		defer logger.Sync()
		Log = logger
	}

	return Log
}
