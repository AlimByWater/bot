package users

import (
	"context"
	"elysium/internal/entity"
	"log/slog"
	"time"
)

func (m *Module) GetOnlineUsersCount() map[string]int64 {
	return m.streamsOnlineCount
}

func (m *Module) GetAllCurrentListeners(ctx context.Context) ([]entity.ListenerCache, error) {
	start := time.Now()
	count, err := m.cache.GetAllCurrentListeners(m.ctx)
	if err != nil {
		m.logger.Error("Failed to get listeners count", slog.String("error", err.Error()), slog.String("method", "GetAllCurrentListeners"))
		return nil, err
	}
	m.logger.Info("PROFILING get all current listeners", slog.Float64("seconds", time.Since(start).Seconds()))

	return count, nil
}

func (m *Module) updateOnlineUsersCountLoop() {
	for {
		start := time.Now()
		count, err := m.cache.GetListenersCount(m.ctx)
		if err != nil {
			m.logger.Error("Failed to get listeners count", slog.String("error", err.Error()), slog.String("method", "UpdateOnlineUsersCount"))
			continue
		}

		m.logger.Info("Get listeners count PROFILING", slog.Float64("seconds", time.Since(start).Seconds()))

		//count = m.alterCount(count)

		m.mu.Lock()
		m.streamsOnlineCount = count
		m.mu.Unlock()
		time.Sleep(5 * time.Second)
	}
}

func (m *Module) alterCount(currentCount int64) int64 {
	return currentCount
}
