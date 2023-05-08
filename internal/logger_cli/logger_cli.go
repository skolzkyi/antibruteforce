package logger_cli

import (
	"go.uber.org/zap"
)

type LogWrap struct {
	config zap.Config
	logger *zap.SugaredLogger
}

func New(level string) (*LogWrap, error) {
	logWrap := LogWrap{}
	zlevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	logWrap.config = zap.Config{
		Level:            zlevel,
		DisableCaller:    true,
		Development:      true,
		Encoding:         "console",
		OutputPaths:      []string{"cli_log.log"},
		ErrorOutputPaths: []string{"cli_log.log"},
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
	}
	logWrap.logger = zap.Must(logWrap.config.Build()).Sugar()
	return &logWrap, nil
}

func (l LogWrap) Info(msg string) {
	l.logger.Info(msg)
}

func (l LogWrap) Error(msg string) {
	l.logger.Error(msg)
}

