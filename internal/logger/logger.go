package logger

import (
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// InitLogger 初始化日志器
func InitLogger(logPath string, level zapcore.Level) error {
	var err error
	once.Do(func() {
		logger, err = createLogger(logPath, level)
	})
	return err
}

// createLogger 创建日志器
func createLogger(logPath string, level zapcore.Level) (*zap.Logger, error) {
	// 创建日志目录
	if logPath != "" {
		dir := filepath.Dir(logPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// 配置编码器（控制台友好格式）
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,  // 彩色大写级别
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO 时间格式
		EncodeDuration: zapcore.StringDurationEncoder,   // 人类可读的持续时间
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短调用者信息
		ConsoleSeparator: "  ",                           // 控制台分隔符
	}

	// 创建编码器
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 创建写入器
	var writers []zapcore.WriteSyncer
	
	// 控制台输出
	writers = append(writers, zapcore.AddSync(os.Stdout))
	
	// 文件输出
	if logPath != "" {
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		writers = append(writers, zapcore.AddSync(file))
	}

	// 创建核心
	core := zapcore.NewCore(encoder, zapcore.NewMultiWriteSyncer(writers...), level)

	// 创建日志器
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)), nil
}

// GetLogger 获取日志器
func GetLogger() *zap.Logger {
	if logger == nil {
		// 如果未初始化，创建默认日志器
		defaultLogger, _ := zap.NewProduction()
		return defaultLogger
	}
	return logger
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 记录致命日志
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Sync 同步日志
func Sync() error {
	return GetLogger().Sync()
}