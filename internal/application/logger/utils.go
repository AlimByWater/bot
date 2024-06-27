package logger

import "log/slog"

func AppendErrorToLogs(oldAttrs []slog.Attr, err error) []slog.Attr {
	return AppendToLogs(oldAttrs, slog.String("err", err.Error()))
}

func AppendToLogs(oldAttrs []slog.Attr, attr slog.Attr) []slog.Attr {
	newAttrs := make([]slog.Attr, 0, len(oldAttrs)+1)
	newAttrs = append(newAttrs, oldAttrs...)
	newAttrs = append(newAttrs, attr)
	return newAttrs
}
