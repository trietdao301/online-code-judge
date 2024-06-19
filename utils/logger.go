package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger() (logger *zap.Logger) {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	fileEncoder := zapcore.NewConsoleEncoder(cfg)
	logWriter := zapcore.Lock(os.Stdout)
	core := zapcore.NewCore(fileEncoder, logWriter, zap.NewAtomicLevelAt(zap.InfoLevel))

	loggerApp := zap.New(core)
	defer loggerApp.Sync()
	return loggerApp
}
