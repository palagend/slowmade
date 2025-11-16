// pkg/logging/logger.go
package logging

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once   sync.Once
	logger *zap.Logger
)

// Config 结构体与您的 TOML 配置匹配
type Config struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"` // console 或 json
	File     string `mapstructure:"file"`     // 对应配置中的 file 字段
}

// Init 初始化日志系统
func Init(config Config) error {
	var initErr error
	once.Do(func() {
		// 设置日志级别
		var level zapcore.Level
		if err := level.UnmarshalText([]byte(config.Level)); err != nil {
			// 默认使用 info 级别
			level = zapcore.InfoLevel
		}

		// 创建编码器配置
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.TimeKey = "time"
		encoderConfig.LevelKey = "level"
		encoderConfig.MessageKey = "message"
		encoderConfig.CallerKey = "caller"
		encoderConfig.StacktraceKey = "stacktrace"
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		// 创建编码器
		var encoder zapcore.Encoder
		switch config.Encoding {
		case "console":
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		case "json", "":
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		default:
			encoder = zapcore.NewJSONEncoder(encoderConfig)
		}

		// 创建输出同步器
		writeSyncer, err := createWriteSyncer(config)
		if err != nil {
			initErr = err
			return
		}

		// 创建核心
		core := zapcore.NewCore(encoder, writeSyncer, level)

		// 创建 logger
		logger = zap.New(core,
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
	})
	return initErr
}

// createWriteSyncer 创建日志输出同步器
func createWriteSyncer(config Config) (zapcore.WriteSyncer, error) {
	var syncers []zapcore.WriteSyncer

	// 总是添加标准错误输出（用于控制台）
	syncers = append(syncers, zapcore.AddSync(os.Stderr))

	// 如果配置了文件输出，添加文件输出
	if config.File != "" {
		// 使用 lumberjack 实现日志轮转
		lumberjackLogger := &lumberjack.Logger{
			Filename:   config.File,
			MaxSize:    100,  // MB
			MaxBackups: 10,   // 保留的旧日志文件最大数量
			MaxAge:     30,   // 保留旧日志文件的最大天数
			Compress:   true, // 压缩旧日志
		}
		syncers = append(syncers, zapcore.AddSync(lumberjackLogger))
	}

	// 使用 MultiWriteSyncer 同时输出到控制台和文件
	return zapcore.NewMultiWriteSyncer(syncers...), nil
}

// Get 获取日志记录器实例
func Get() *zap.Logger {
	if logger == nil {
		// 如果未初始化，使用默认配置初始化
		defaultConfig := Config{
			Level:    "info",
			Encoding: "console",
			File:     "slowmade.log",
		}
		if err := Init(defaultConfig); err != nil {
			panic("logger initialization failed: " + err.Error())
		}
	}
	return logger
}

// Sync 刷新日志缓冲区
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

// 快捷方法 - 结构化日志
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

// 快捷方法 - 格式化日志（SugaredLogger）
func Sugar() *zap.SugaredLogger {
	return Get().Sugar()
}

func Debugf(template string, args ...interface{}) {
	Sugar().Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	Sugar().Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	Sugar().Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	Sugar().Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	Sugar().Fatalf(template, args...)
}
