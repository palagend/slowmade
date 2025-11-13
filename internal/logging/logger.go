package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/palagend/slowmade/internal/config"
)

// 日志级别
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

var levelStrings = map[Level]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
	PanicLevel: "PANIC",
}

var stringToLevel = map[string]Level{
	"debug": DebugLevel,
	"info":  InfoLevel,
	"warn":  WarnLevel,
	"error": ErrorLevel,
	"fatal": FatalLevel,
	"panic": PanicLevel,
}

// 日志条目
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Caller    string                 `json:"caller,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// 全局日志器
type Logger struct {
	config     *config.LoggingConfig
	level      Level
	output     io.Writer
	mu         sync.Mutex
	fields     map[string]interface{}
	timeFormat string
}

var (
	globalLogger *Logger
	once         sync.Once
)

// 初始化全局日志器
func Initialize(cfg *config.LoggingConfig) error {
	var err error
	once.Do(func() {
		globalLogger, err = NewLogger(cfg)
	})
	return err
}

// 获取全局日志器实例
func Global() *Logger {
	if globalLogger == nil {
		// 返回一个默认的日志器
		cfg := &config.LoggingConfig{
			Level:  "info",
			Format: "console",
			Output: "stderr",
		}
		globalLogger, _ = NewLogger(cfg)
	}
	return globalLogger
}

// 创建新的日志器
func NewLogger(cfg *config.LoggingConfig) (*Logger, error) {
	logger := &Logger{
		config:     cfg,
		fields:     make(map[string]interface{}),
		timeFormat: "2006-01-02 15:04:05",
	}

	// 设置日志级别
	level, ok := stringToLevel[strings.ToLower(cfg.Level)]
	if !ok {
		level = InfoLevel // 默认级别
	}
	logger.level = level

	// 设置时间格式
	if cfg.TimeFormat != "" {
		logger.timeFormat = cfg.TimeFormat
	}

	// 设置输出目标
	var output io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stdout":
		output = os.Stdout
	case "stderr", "":
		output = os.Stderr
	default:
		// 文件输出
		dir := filepath.Dir(cfg.Output)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %v", err)
		}
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("打开日志文件失败: %v", err)
		}
		output = file
	}
	logger.output = output

	return logger, nil
}

// 带字段的日志方法
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := *l
	newLogger.fields = make(map[string]interface{})
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	for k, v := range fields {
		newLogger.fields[k] = v
	}
	return &newLogger
}

// 获取调用者信息
func (l *Logger) getCaller() string {
	if !l.config.EnableCaller {
		return ""
	}
	_, file, line, ok := runtime.Caller(3) // 跳过3层调用栈
	if ok {
		return fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}
	return ""
}

// 核心日志方法
func (l *Logger) log(level Level, msg string, fields map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     levelStrings[level],
		Message:   msg,
		Caller:    l.getCaller(),
	}

	// 合并字段
	allFields := make(map[string]interface{})
	for k, v := range l.fields {
		allFields[k] = v
	}
	for k, v := range fields {
		allFields[k] = v
	}
	if len(allFields) > 0 {
		entry.Fields = allFields
	}

	// 堆栈跟踪
	if level >= ErrorLevel && l.config.EnableStack {
		buf := make([]byte, 1024)
		n := runtime.Stack(buf, false)
		entry.Stack = string(buf[:n])
	}

	l.writeEntry(entry)

	// 严重错误处理
	if level == FatalLevel {
		os.Exit(1)
	} else if level == PanicLevel {
		panic(msg)
	}
}

// 写入日志条目
func (l *Logger) writeEntry(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var logLine string
	switch strings.ToLower(l.config.Format) {
	case "json":
		data, _ := json.Marshal(entry)
		logLine = string(data) + "\n"
	default:
		// 控制台格式
		logLine = l.formatConsole(entry)
	}

	fmt.Fprint(l.output, logLine)
}

// 控制台格式格式化
func (l *Logger) formatConsole(entry LogEntry) string {
	var builder strings.Builder

	// 时间戳
	builder.WriteString(entry.Timestamp.Format(l.timeFormat))
	builder.WriteString(" ")

	// 级别（带颜色）
	levelStr := entry.Level
	if l.config.Color {
		levelStr = l.colorizeLevel(entry.Level)
	}
	builder.WriteString(levelStr)
	builder.WriteString(" ")

	// 调用者信息
	if entry.Caller != "" {
		builder.WriteString("[")
		builder.WriteString(entry.Caller)
		builder.WriteString("] ")
	}

	// 消息
	builder.WriteString(entry.Message)

	// 字段
	if len(entry.Fields) > 0 {
		builder.WriteString(" {")
		first := true
		for k, v := range entry.Fields {
			if !first {
				builder.WriteString(", ")
			}
			builder.WriteString(k)
			builder.WriteString("=")
			builder.WriteString(fmt.Sprintf("%v", v))
			first = false
		}
		builder.WriteString("}")
	}

	builder.WriteString("\n")
	return builder.String()
}

// 级别颜色化
func (l *Logger) colorizeLevel(level string) string {
	if !l.config.Color {
		return level
	}

	colorCodes := map[string]string{
		"DEBUG": "\033[36m", // 青色
		"INFO":  "\033[32m", // 绿色
		"WARN":  "\033[33m", // 黄色
		"ERROR": "\033[31m", // 红色
		"FATAL": "\033[35m", // 洋红色
		"PANIC": "\033[35m", // 洋红色
	}

	reset := "\033[0m"
	if color, exists := colorCodes[level]; exists {
		return color + level + reset
	}
	return level
}

// 公共日志方法
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.log(DebugLevel, msg, mergeFields(fields...))
}

func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.log(InfoLevel, msg, mergeFields(fields...))
}

func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.log(WarnLevel, msg, mergeFields(fields...))
}

func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	l.log(ErrorLevel, msg, mergeFields(fields...))
}

func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {

	l.log(FatalLevel, msg, mergeFields(fields...))
}

func (l *Logger) Panic(msg string, fields ...map[string]interface{}) {
	l.log(PanicLevel, msg, mergeFields(fields...))
}

// 全局快捷方法
func Debug(msg string, fields ...map[string]interface{}) {
	Global().Debug(msg, fields...)
}

func Info(msg string, fields ...map[string]interface{}) {
	Global().Info(msg, fields...)
}

func Warn(msg string, fields ...map[string]interface{}) {
	Global().Warn(msg, fields...)
}

func Error(msg string, fields ...map[string]interface{}) {
	Global().Error(msg, fields...)
}

func Fatal(msg string, fields ...map[string]interface{}) {
	Global().Fatal(msg, fields...)
}

// 辅助函数
func mergeFields(fields ...map[string]interface{}) map[string]interface{} {
	if len(fields) == 0 {
		return nil
	}
	return fields[0]
}
