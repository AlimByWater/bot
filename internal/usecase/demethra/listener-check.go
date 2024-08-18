package demethra

import (
	"log/slog"
	"time"
)

func (m *Module) CheckIdleListenersAndSaveListeningDuration() {
	listeners, err := m.cache.GetAllCurrentListeners(m.ctx)
	if err != nil {
		m.logger.Error("Failed to get listeners", slog.String("error", err.Error()), slog.String("method", "CheckIdleListenersAndSaveListeningDuration"))
		return
	}

	for _, listener := range listeners {
		// 12313451 - 12313250 = 201    3 * 60 = 180
		if time.Now().Unix()-listener.Payload.LastActivity > int64(m.cfg.GetListenerIdleTimeoutInMinutes())*60 {
			err = m.cache.RemoveListenerTelegramID(m.ctx, listener.TelegramID)
			if err != nil {
				m.logger.Error("Failed to delete listener", slog.Int64("telegram_id", listener.TelegramID), slog.String("error", err.Error()), slog.String("method", "CheckIdleListenersAndSaveListeningDuration"))
			}

			continue
		}
	}
}
