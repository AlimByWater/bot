package telegram

import (
	"fmt"
	"log/slog"
)

type telegoLogger struct {
	logger *slog.Logger
}

func (l telegoLogger) Debugf(format string, args ...any) {
	l.logger.Debug(fmt.Sprintf(format+"\n", args...))
}

func (l telegoLogger) Errorf(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format+"\n", args...))
}
