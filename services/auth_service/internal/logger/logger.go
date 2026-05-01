package logger

import (
	"go.uber.org/zap"
)

type Logger struct {
	HttpLogger     *zap.Logger
	ServerLogger   *zap.Logger
	ProducerLogger *zap.Logger
}

func NewLogger() *Logger {
	loggerCfg := zap.NewProductionConfig()
	loggerCfg.OutputPaths = []string{"stdout"}
	logger, err := loggerCfg.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{
		HttpLogger:     logger.With(zap.String("layer", "http")),
		ServerLogger:   logger.With(zap.String("layer", "server")),
		ProducerLogger: logger.With(zap.String("layer", "producer")),
	}
}
