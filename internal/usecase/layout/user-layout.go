package layout

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"log/slog"
	"time"
)

func (m *Module) GetLayoutByName(ctx context.Context, layoutName string, initiatorUserID int) (entity.UserLayout, error) {
	// Попытка получить макет из кэша
	cachedLayout, err := m.cache.GetLayoutByName(ctx, layoutName)
	if err != nil {
		m.logger.Debug("cache GetLayoutByName", slog.String("layout name", layoutName), slog.String("error", err.Error()), slog.String("method", "GetLayoutByName"))
	} else if cachedLayout.ID != 0 {
		return m.filterLayout(cachedLayout, initiatorUserID), nil
	}

	// Если макет не найден в кэше, получаем его из репозитория
	layout, err := m.repo.LayoutByName(ctx, layoutName)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("repo LayoutByName: %w", err)
	}

	// Сохраняем макет в кэш
	err = m.cache.SaveOrUpdateLayout(ctx, layout)
	if err != nil {
		m.logger.Error("Не удалось сохранить макет в кэш", "error", err)
	}

	return m.filterLayout(layout, initiatorUserID), nil
}

// GetUserLayout получает макет пользователя с учетом прав доступа инициатора
func (m *Module) GetUserLayout(ctx context.Context, userID, initiatorUserID int) (entity.UserLayout, error) {
	// Попытка получить макет из кэша
	cachedLayout, err := m.cache.GetLayoutByUserID(ctx, userID)
	if err != nil {
		m.logger.Debug("cache GetLayoutByUserID", slog.Int("userID", userID), slog.String("error", err.Error()), slog.String("method", "GetUserLayout"))
	} else if cachedLayout.ID != 0 {
		return m.filterLayout(cachedLayout, initiatorUserID), nil
	}

	// Если макет не найден в кэше, получаем его из репозитория
	layout, err := m.repo.LayoutByUserID(ctx, userID)
	if err != nil {
		if err == entity.ErrLayoutNotFound {
			layout, err = m.GenerateAndSaveDefaultLayout(ctx, userID, "")
			if err != nil {
				return entity.UserLayout{}, fmt.Errorf("GenerateAndSaveDefaultLayout: %w", err)
			}
		} else {
			return entity.UserLayout{}, fmt.Errorf("repo LayoutByUserID: %w", err)
		}
	}

	// Сохраняем макет в кэш
	err = m.cache.SaveOrUpdateLayout(ctx, layout)
	if err != nil {
		m.logger.Error("Не удалось сохранить макет в кэш", "error", err)
	}

	return m.filterLayout(layout, initiatorUserID), nil
}

// filterLayout фильтрует элементы макета в зависимости от прав доступа
func (m *Module) filterLayout(layout entity.UserLayout, initiatorUserID int) entity.UserLayout {
	if !m.hasEditPermission(layout, initiatorUserID) {
		filteredElements := make([]entity.LayoutElement, 0, len(layout.Elements))
		for _, element := range layout.Elements {
			if element.IsPublic {
				filteredElements = append(filteredElements, element)
			}
		}
		layout.Elements = filteredElements
	}
	return layout
}

// UpdateLayoutFull обновляет макет пользователя полностью
func (m *Module) UpdateLayoutFull(ctx context.Context, layoutID int, initiatorUserID int, updatedLayout entity.UserLayout) error {
	currentLayout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return fmt.Errorf("не удалось получить текущий макет: %w", err)
	}

	if !m.hasEditPermission(currentLayout, initiatorUserID) {
		return entity.ErrNoPermissionToEditLayout
	}

	updatedLayout.Creator = currentLayout.Creator
	updatedLayout.Editors = currentLayout.Editors
	updatedLayout.ID = layoutID

	err = m.repo.UpdateLayoutFull(ctx, updatedLayout)
	if err != nil {
		return fmt.Errorf("не удалось обновить макет: %w", err)
	}

	// Обновляем кэш
	err = m.cache.SaveOrUpdateLayout(ctx, updatedLayout)
	if err != nil {
		m.logger.Error("Не удалось обновить макет в кэше", "error", err)
	}

	err = m.logLayoutChange(ctx, initiatorUserID, layoutID, "UpdateLayoutFull", nil)
	if err != nil {
		m.logger.Error("Не удалось записать изменение макета", "error", err)
	}
	return nil
}

// GetLayout получает макет по его ID с учетом прав доступа инициатора
func (m *Module) GetLayout(ctx context.Context, layoutID int, initiatorUserID int) (entity.UserLayout, error) {
	// Попытка получить макет из кэша
	cachedLayout, err := m.cache.GetLayout(ctx, layoutID)
	if err != nil {
		m.logger.Debug("cache GetLayout", slog.Int("layoutID", layoutID), slog.String("error", err.Error()), slog.String("method", "GetLayout"))
	} else if cachedLayout.ID != 0 {
		return m.filterLayout(cachedLayout, initiatorUserID), nil
	}

	// Если макет не найден в кэше, получаем его из репозитория
	layout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("repo LayoutByID: %w", err)
	}

	// Сохраняем макет в кэш
	err = m.cache.SaveOrUpdateLayout(ctx, layout)
	if err != nil {
		m.logger.Error("Не удалось сохранить макет в кэш", "error", err)
	}

	return m.filterLayout(layout, initiatorUserID), nil
}

// IsEditor проверяет, является ли пользователь редактором макета
func (m *Module) IsEditor(ctx context.Context, layoutID int, userID int) (bool, error) {
	layout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return false, fmt.Errorf("не удалось получить макет: %w", err)
	}

	return m.hasEditPermission(layout, userID), nil
}

// hasEditPermission проверяет, имеет ли пользователь права на редактирование макета
func (m *Module) hasEditPermission(layout entity.UserLayout, userID int) bool {
	if layout.Creator == userID {
		return true
	}
	for _, editor := range layout.Editors {
		if editor == userID {
			return true
		}
	}
	return false
}

// logLayoutChange записывает изменение макета в лог
func (m *Module) logLayoutChange(ctx context.Context, userID int, layoutID int, changeType string, details map[string]interface{}) error {
	change := entity.LayoutChange{
		UserID:     userID,
		LayoutID:   layoutID,
		Timestamp:  time.Now(),
		ChangeType: changeType,
		Details:    details,
	}
	err := m.repo.LogLayoutChange(ctx, change)
	if err != nil {
		m.logger.Error("Не удалось записать изменение макета",
			"error", err,
			"userID", change.UserID,
			"layoutID", change.LayoutID,
			"change_type", change.ChangeType,
			"details", change.Details)
		return fmt.Errorf("не удалось записать изменение макета: %w", err)
	}
	return nil
}

// GenerateAndSaveDefaultLayout генерирует и сохраняет стандартный макет для пользователя
func (m *Module) GenerateAndSaveDefaultLayout(ctx context.Context, userID int, username string) (entity.UserLayout, error) {
	defaultLayout, err := m.repo.GetDefaultLayout(ctx)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("не удалось получить стандартный макет: %w", err)
	}
	defaultLayout.Creator = userID
	defaultLayout.Name = fmt.Sprintf("%s-%d", username, userID)
	err = m.repo.CreateLayout(ctx, defaultLayout)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("не удалось сохранить стандартный макет: %w", err)
	}
	return defaultLayout, nil
}
