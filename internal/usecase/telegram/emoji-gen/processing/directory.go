package processing

import (
	"elysium/internal/entity"
	"fmt"
	"log/slog"
	"os"
	"time"
)

func (m *Module) RegisterDirectory(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	m.directories.Store(dir, struct{}{})
	return nil
}

func (m *Module) CheckAndRemoveOldDirectories() {
	now := time.Now()
	m.directories.Range(func(key, value interface{}) bool {
		dir := key.(string)
		defer m.directories.Delete(key)

		// Удаляем старые директории
		// Проверяем, как давно создана директория
		// Если создана больше 7 дней, то удаляем

		info, err := os.Stat(dir)
		if err != nil {
			slog.Error("Failed to stat directory", slog.String("dir", dir), slog.String("err", err.Error()))
			return true
		}

		remove := info.ModTime().Add(entity.DirectoryLifeTime).Before(now)
		if remove {
			slog.Info("Removing old directory", slog.String("dir", dir))
			if err := os.RemoveAll(dir); err != nil {
				slog.Error("Failed to remove directory", slog.String("dir", dir), slog.String("err", err.Error()))
				return true
			}
		}

		return true

	})
}
