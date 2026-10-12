package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"recon/internal/config"
)

var Logger *zap.Logger

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

func init() {

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(config.GlobalConfig.Log.Level)); err != nil {
		fmt.Println(err)
		return
	}

	location, err := time.LoadLocation(config.GlobalConfig.System.Timezone)
	if err != nil {

		location = time.Local
		fmt.Println("load Timezone error")
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalColorLevelEncoder,
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {

			t = t.In(location)
			encoder.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	var cores []zapcore.Core

	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
	cores = append(cores, consoleCore)

	if config.GlobalConfig.Log.Filename != "" {
		logDir := filepath.Dir(config.GlobalConfig.Log.Filename)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			fmt.Println(err)
			return
		}

		fileWriter := &lumberjack.Logger{
			Filename: config.GlobalConfig.Log.Filename,
			MaxSize:  300,
			Compress: true,
		}

		fileCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(fileWriter), level)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)

	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return
}
