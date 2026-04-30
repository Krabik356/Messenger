package logger

import "go.uber.org/zap"

type Logger struct {
	HttpLogger   *zap.Logger
	ServerLogger *zap.Logger
}

func createLoggerWithLayer(layer string) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{"stdout"}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	logger.With(zap.String("layer", layer))
	return logger
}

func NewLogger() *Logger {
	return &Logger{
		HttpLogger:   createLoggerWithLayer("http"),
		ServerLogger: createLoggerWithLayer("server"),
	}
}
