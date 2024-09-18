package demethra

import (
	"context"
	"elysium/internal/entity"
	"log/slog"
)

func (m *Module) ProcessWebAppState(ctx context.Context, state entity.WebAppState) {
	err := m.cache.SaveOrUpdateListener(ctx, entity.ListenerCache{
		TelegramID: state.TelegramID,
		Payload:    entity.ListenerCachePayload{}, // метод SaveOrUpdateListener при пустом payload обновит lastActivity
	})

	if err != nil {
		m.logger.Error("Failed to save or update listener", slog.String("error", err.Error()), slog.Int64("telegram_id", state.TelegramID), slog.String("method", "ProcessWebAppState"))
	}
}
