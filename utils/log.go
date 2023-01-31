package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func init() {
	InitLog()
}

var lg *zap.SugaredLogger

func Log() *zap.SugaredLogger {
	return lg
}

func InitLog() *zap.SugaredLogger {
	consoleConfig := setConfigs(zap.NewProductionEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)
	core := zapcore.NewTee(zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	lg = logger.Sugar()
	return lg
}

// Google Cloud Platform doesn't want file logs, so this is not used
func InitLogWithFileLog(filename string) *zap.SugaredLogger {
	defaultLogLevel := zapcore.DebugLevel

	consoleConfig := setConfigs(zap.NewProductionEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(consoleConfig)

	fileConfig := setConfigs(zap.NewProductionEncoderConfig())
	fileEncoder := zapcore.NewJSONEncoder(fileConfig)

	logFile := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    1, // megabytes
		MaxBackups: 2,
		MaxAge:     1, // days
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), defaultLogLevel),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), defaultLogLevel),
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	lg = logger.Sugar()
	return lg
}

func setConfigs(cfg zapcore.EncoderConfig) zapcore.EncoderConfig {
	cfg.EncodeLevel = encodeLevel()
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout(TimestampFormatLayout)
	cfg.CallerKey = "c"
	cfg.LevelKey = "l"
	cfg.MessageKey = "m"
	// cfg.StacktraceKey = stacktrace // Use "" to disable stacktrace
	cfg.TimeKey = "t"
	return cfg
}

// Zap and GCP: https://github.com/uber-go/zap/issues/1095
func encodeLevel() zapcore.LevelEncoder {
	return func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.DebugLevel:
			enc.AppendString("DEBUG")
		case zapcore.InfoLevel:
			enc.AppendString("INFO")
		case zapcore.WarnLevel:
			enc.AppendString("WARNING")
		case zapcore.ErrorLevel:
			enc.AppendString("ERROR")
		case zapcore.DPanicLevel:
			enc.AppendString("CRITICAL")
		case zapcore.PanicLevel:
			enc.AppendString("ALERT")
		case zapcore.FatalLevel:
			enc.AppendString("EMERGENCY")
		}
	}
}
