package logger

import (
	"io"
	"log/slog"
	"os"
	"strconv"
)

var defaultHandlerOptions = &slog.HandlerOptions{
	AddSource: false,
	Level:     slog.LevelInfo,
}

// Options структура конфигов для логгера
type Options struct {
	AppName        string
	Writer         io.Writer
	HandlerOptions *slog.HandlerOptions
}

// New конструктор логгера
func New(options Options) *slog.Logger {
	if options.HandlerOptions == nil {
		options.HandlerOptions = defaultHandlerOptions

		if sourceEnv, ok := os.LookupEnv("LOGGER_SOURCE"); ok {
			if source, err := strconv.ParseBool(sourceEnv); err == nil {
				options.HandlerOptions.AddSource = source
			}
		}
		if levelEnv, ok := os.LookupEnv("LOGGER_LEVEL"); ok {
			if level, err := strconv.Atoi(levelEnv); err == nil {
				options.HandlerOptions.Level = slog.Level(level)
			}
		}
	}
	return slog.New(
		slog.NewJSONHandler(
			options.Writer,
			options.HandlerOptions,
		),
	).With(slog.String("app", options.AppName))
}
