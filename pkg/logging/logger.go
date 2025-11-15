// pkg/logging/logger.go
package logging

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once   sync.Once
	logger *zap.Logger
)

type Config struct {
	Level      string `mapstructure:"level"`
	Encoding   string `mapstructure:"encoding"`
	OutputPath string `mapstructure:"output_path"`
}

func Init(config Config) error {
	var initErr error
	once.Do(func() {
		var level zapcore.Level
		if err := level.UnmarshalText([]byte(config.Level)); err != nil {
			initErr = err
			return
		}

		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stderr),
			level,
		)

		if config.OutputPath != "" {
			file, err := os.OpenFile(config.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				initErr = err
				return
			}
			core = zapcore.NewTee(
				core,
				zapcore.NewCore(
					zapcore.NewJSONEncoder(encoderConfig),
					zapcore.AddSync(file),
					level,
				),
			)
		}

		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	})
	return initErr
}

func Get() *zap.Logger {
	if logger == nil {
		panic("logger not initialized")
	}
	return logger
}

func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}
