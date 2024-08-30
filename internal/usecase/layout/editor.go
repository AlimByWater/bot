package layout

import (
	"context"
	"errors"
	"fmt"
)

// AddLayoutEditor добавляет редактора к макету
func (m *Module) AddLayoutEditor(ctx context.Context, layoutID string, initiatorUserID, editorID int) error {
	// Проверяем, имеет ли инициатор права на редактирование макета
	isOwner, err := m.repo.IsLayoutOwner(ctx, layoutID, initiatorUserID)
	if err != nil {
		return fmt.Errorf("failed to check layout ownership: %w", err)
	}
	if !isOwner {
		return errors.New("initiator does not have permission to add editors")
	}

	// Добавляем нового редактора
	err = m.repo.AddLayoutEditor(ctx, layoutID, editorID)
	if err != nil {
		return fmt.Errorf("failed to add layout editor: %w", err)
	}

	// Инвалидируем кэш для данного макета
	cacheKey := fmt.Sprintf("layout:%s", layoutID)
	err = m.cache.Delete(ctx, cacheKey)
	if err != nil {
		m.logger.Error("Failed to invalidate layout cache", "error", err)
	}

	// Логируем изменение
	err = m.logLayoutChange(ctx, initiatorUserID, layoutID, "AddLayoutEditor", fmt.Sprintf("Added editor %d", editorID))
	if err != nil {
		m.logger.Error("Failed to log layout change", "error", err)
	}

	return nil
}

// RemoveLayoutEditor удаляет редактора из макета
func (m *Module) RemoveLayoutEditor(ctx context.Context, layoutID string, initiatorUserID, editorID int) error {
	// Проверяем, имеет ли инициатор права на редактирование макета
	isOwner, err := m.repo.IsLayoutOwner(ctx, layoutID, initiatorUserID)
	if err != nil {
		return fmt.Errorf("failed to check layout ownership: %w", err)
	}
	if !isOwner {
		return errors.New("initiator does not have permission to remove editors")
	}

	// Удаляем редактора
	err = m.repo.RemoveLayoutEditor(ctx, layoutID, editorID)
	if err != nil {
		return fmt.Errorf("failed to remove layout editor: %w", err)
	}

	// Инвалидируем кэш для данного макета
	cacheKey := fmt.Sprintf("layout:%s", layoutID)
	err = m.cache.Delete(ctx, cacheKey)
	if err != nil {
		m.logger.Error("Failed to invalidate layout cache", "error", err)
	}

	// Логируем изменение
	err = m.logLayoutChange(ctx, initiatorUserID, layoutID, "RemoveLayoutEditor", fmt.Sprintf("Removed editor %d", editorID))
	if err != nil {
		m.logger.Error("Failed to log layout change", "error", err)
	}

	return nil
}
