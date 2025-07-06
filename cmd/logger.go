package cmd

import (
	"log/slog"
	"os"

	"github.com/Easy-Infra-Ltd/easy-test/internal/logger"
)

type LoggerOptions struct {
	Verbose   bool
	NoColor   bool
	Level     slog.Level
	ColorName string
}

func CreateLoggerOptions(verbose, noColor bool, logLevel, colorName string) LoggerOptions {
	return LoggerOptions{
		Verbose:   verbose,
		NoColor:   noColor,
		Level:     parseLogLevel(logLevel),
		ColorName: colorName,
	}
}

func CreateLogger(opts LoggerOptions) *slog.Logger {
	if opts.NoColor {
		return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: opts.Level,
		}))
	}

	return logger.CreateLoggerFromEnv(nil, opts.ColorName)
}

func ConfigureProcessLogger(baseLogger *slog.Logger, processName, area string, dryRun bool) *slog.Logger {
	if dryRun {
		processName = "[DRY]" + processName
	}
	
	configuredLogger := baseLogger.With("process", processName).With("area", area)
	slog.SetDefault(configuredLogger)
	
	return configuredLogger
}