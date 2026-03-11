package initialize

import (
	"bookadmin/global"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Zap() *zap.Logger {
	writerSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writerSyncer, getLogLevel())
	logger := zap.New(core, zap.AddCaller())
	global.GVA_LOG = logger
	zap.ReplaceGlobals(logger)
	return logger
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	if global.GVA_CONFIG != nil && global.GVA_CONFIG.IsProduction() {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	logFile := "./log/app.log"
	maxSize := 10
	maxBackups := 5
	maxAge := 30
	compress := false

	if global.GVA_CONFIG != nil {
		logFile = global.GVA_CONFIG.Log.File
		maxSize = global.GVA_CONFIG.Log.MaxSize
		maxBackups = global.GVA_CONFIG.Log.MaxBackups
		maxAge = global.GVA_CONFIG.Log.MaxAge
		compress = global.GVA_CONFIG.Log.Compress
	}

	if dir := filepath.Dir(logFile); dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}

	lumberJackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}
	return zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberJackLogger))
}

func getLogLevel() zapcore.Level {
	if global.GVA_CONFIG == nil {
		return zapcore.InfoLevel
	}

	switch strings.ToLower(global.GVA_CONFIG.Log.Level) {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
