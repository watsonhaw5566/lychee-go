package logger

import (
	"os"
	"time"

	"lychee-go/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.SugaredLogger

// Init 初始化日志系统
func Init() {
	logFile := config.GetString("log.filename", "runtime/logs/app.log")
	logLevel := config.GetString("log.level", "info")
	maxSize := config.GetInt("log.max_size", 100)
	maxBackups := config.GetInt("log.max_backups", 30)
	maxAge := config.GetInt("log.max_age", 7)
	compress := config.GetBool("log.compress", true)

	_ = os.MkdirAll("runtime/logs", 0755)

	var zapLevel zapcore.Level
	switch logLevel {
	case "debug", "DEBUG":
		zapLevel = zapcore.DebugLevel
	case "warn", "WARN":
		zapLevel = zapcore.WarnLevel
	case "error", "ERROR":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05"))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	})
	fileEncoderConfig := encoderConfig
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, zapLevel)

	core := zapcore.NewTee(consoleCore, fileCore)

	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	Log = logger.Sugar()
	Log.Info("Logger initialized successfully")
}

// Debug 调试级别
func Debug(msg string, args ...interface{}) {
	Log.Debugf(msg, args...)
}

// Info 信息级别
func Info(msg string, args ...interface{}) {
	Log.Infof(msg, args...)
}

// Warn 警告级别
func Warn(msg string, args ...interface{}) {
	Log.Warnf(msg, args...)
}

// Error 错误级别
func Error(msg string, args ...interface{}) {
	Log.Errorf(msg, args...)
}

// Fatal 致命错误
func Fatal(msg string, args ...interface{}) {
	Log.Fatalf(msg, args...)
}

// With 携带上下文（结构化日志）
func With(args ...interface{}) *zap.SugaredLogger {
	return Log.With(args...)
}

// Sync 刷新日志缓冲
func Sync() {
	_ = Log.Sync()
}
