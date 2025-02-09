package telegram

import (
	"fmt"
	"log/slog"
	"runtime/debug"
)

type telegoRecovery struct {
	logger *slog.Logger
}

func (r telegoRecovery) Handler(recovered any) {
	if recovered != nil {
		r.logger.Error("panic", slog.Any("err", recovered))
		fmt.Println(string(debug.Stack()))
		// r.logger.Error("panic", slog.Any("err", recovered), slog.String("stack", string(debug.Stack())))
	}
}
