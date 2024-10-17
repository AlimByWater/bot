package users

import (
	"context"
	"elysium/internal/entity"
	"log/slog"
	"time"
)

func (m *Module) GetOnlineUsersCount() int64 {
	return m.onlineUsersCount.Load()
}

func (m *Module) GetAllCurrentListeners(ctx context.Context) ([]entity.ListenerCache, error) {
	return m.cache.GetAllCurrentListeners(ctx)
}

func (m *Module) updateOnlineUsersCountLoop() {
	for {
		count, err := m.cache.GetListenersCount(m.ctx)
		if err != nil {
			m.logger.Error("Failed to get listeners count", slog.String("error", err.Error()), slog.String("method", "UpdateOnlineUsersCount"))
			continue
		}

		//count = m.alterCount(count)

		m.onlineUsersCount.Store(count)
		time.Sleep(5 * time.Second)
	}
}

func (m *Module) alterCount(currentCount int64) int64 {
	return currentCount
}
