package logger

import (
	"fmt"

	internalconf "github.com/Raschudesny/otus_go_homeworks/hw12_13_14_15_calendar/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(loggerConfig internalconf.LoggerConfig) error {
	config := zap.NewProductionConfig()

	config.EncoderConfig.EncodeDuration = zapcore.MillisDurationEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.ErrorOutputPaths = []string{"stderr"}
	config.OutputPaths = []string{loggerConfig.File, "stdout"}
	if err := config.Level.UnmarshalText([]byte(loggerConfig.Level)); err != nil {
		return fmt.Errorf("error building logger, can't parse level config value: %w", err)
	}

	currentLogger, err := config.Build()
	if err != nil {
		return fmt.Errorf("error building logger: %w", err)
	}
	zap.ReplaceGlobals(currentLogger)
	return nil
}
