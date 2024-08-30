package layout

import (
	"arimadj-helper/internal/entity"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// GetUserLayout получает макет пользователя с учетом прав доступа инициатора
func (m *Module) GetUserLayout(ctx context.Context, userID, initiatorUserID int) (entity.UserLayout, error) {
	cacheKey := fmt.Sprintf("user_layout:%d", userID)
	
	// Попытка получить макет из кэша
	cachedLayout, err := m.cache.GetFromCache(ctx, cacheKey)
	if err == nil {
		layout, ok := cachedLayout.(entity.UserLayout)
		if ok {
			if !m.hasViewPermission(layout, initiatorUserID) {
				return entity.UserLayout{}, entity.ErrNoPermission
			}
			return m.filterLayout(layout, initiatorUserID), nil
		}
	}

	// Если макет не найден в кэше, получаем его из репозитория
	layout, err := m.repo.LayoutByUserID(ctx, userID)
	if err != nil {
		if err == entity.ErrLayoutNotFound {
			layout, err = m.GenerateAndSaveDefaultLayout(ctx, userID)
			if err != nil {
				return entity.UserLayout{}, err
			}
		} else {
			return entity.UserLayout{}, fmt.Errorf("не удалось получить макет пользователя: %w", err)
		}
	}

	if !m.hasViewPermission(layout, initiatorUserID) {
		return entity.UserLayout{}, entity.ErrNoPermission
	}

	// Сохраняем макет в кэш
	err = m.cache.SetInCache(ctx, cacheKey, layout, 30*time.Minute)
	if err != nil {
		m.logger.Error("Не удалось сохранить макет в кэш", "error", err)
	}

	return m.filterLayout(layout, initiatorUserID), nil
}

// filterLayout фильтрует элементы макета в зависимости от прав доступа
func (m *Module) filterLayout(layout entity.UserLayout, initiatorUserID int) entity.UserLayout {
	if !m.hasEditPermission(layout, initiatorUserID) {
		filteredElements := make([]entity.LayoutElement, 0, len(layout.Layout))
		for _, element := range layout.Layout {
			if element.Public {
				filteredElements = append(filteredElements, element)
			}
		}
		layout.Layout = filteredElements
	}
	return layout
}

// UpdateLayoutFull обновляет макет пользователя полностью
func (m *Module) UpdateLayoutFull(ctx context.Context, layoutID string, initiatorUserID int, updatedLayout entity.UserLayout) error {
	currentLayout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return fmt.Errorf("не удалось получить текущий макет: %w", err)
	}

	if !m.hasEditPermission(currentLayout, initiatorUserID) {
		return entity.ErrNoPermission
	}

	updatedLayout.Creator = currentLayout.Creator
	updatedLayout.Editors = currentLayout.Editors
	updatedLayout.LayoutID = layoutID

	err = m.repo.UpdateLayout(ctx, updatedLayout)
	if err != nil {
		return fmt.Errorf("не удалось обновить макет: %w", err)
	}

	// Обновляем кэш
	cacheKey := fmt.Sprintf("user_layout:%s", updatedLayout.UserID)
	err = m.cache.SetInCache(ctx, cacheKey, updatedLayout, 30*time.Minute)
	if err != nil {
		m.logger.Error("Не удалось обновить макет в кэше", "error", err)
	}

	err = m.logLayoutChange(ctx, initiatorUserID, layoutID, "UpdateLayoutFull", fmt.Sprintf("Макет обновлен"))
	if err != nil {
		m.logger.Error("Не удалось записать изменение макета", "error", err)
	}
	return nil
}

// GetLayout получает макет по его ID с учетом прав доступа инициатора
func (m *Module) GetLayout(ctx context.Context, layoutID string, initiatorUserID int) (entity.UserLayout, error) {
	cacheKey := fmt.Sprintf("layout:%s", layoutID)
	
	// Попытка получить макет из кэша
	cachedLayout, err := m.cache.GetFromCache(ctx, cacheKey)
	if err == nil {
		layout, ok := cachedLayout.(entity.UserLayout)
		if ok {
			return m.filterLayout(layout, initiatorUserID), nil
		}
	}

	// Если макет не найден в кэше, получаем его из репозитория
	layout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("не удалось получить макет: %w", err)
	}

	// Сохраняем макет в кэш
	err = m.cache.SetInCache(ctx, cacheKey, layout, 30*time.Minute)
	if err != nil {
		m.logger.Error("Не удалось сохранить макет в кэш", "error", err)
	}

	return m.filterLayout(layout, initiatorUserID), nil
}

// IsEditor проверяет, является ли пользователь редактором макета
func (m *Module) IsEditor(ctx context.Context, layoutID string, userID int) (bool, error) {
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

// hasViewPermission проверяет, имеет ли пользователь права на просмотр макета
func (m *Module) hasViewPermission(layout entity.UserLayout, userID int) bool {
	return layout.Creator == userID || m.hasEditPermission(layout, userID)
}

// logLayoutChange записывает изменение макета в лог
func (m *Module) logLayoutChange(ctx context.Context, userID int, layoutID string, action, details string) error {
	change := entity.LayoutChange{
		UserID:    userID,
		LayoutID:  layoutID,
		Timestamp: time.Now(),
		Action:    action,
		Details:   details,
	}
	err := m.repo.LogLayoutChange(ctx, change)
	if err != nil {
		m.logger.Error("Не удалось записать изменение макета",
			"error", err,
			"userID", change.UserID,
			"layoutID", change.LayoutID,
			"action", change.Action,
			"details", change.Details)
		return fmt.Errorf("не удалось записать изменение макета: %w", err)
	}
	return nil
}

// GenerateAndSaveDefaultLayout генерирует и сохраняет стандартный макет для пользователя
func (m *Module) GenerateAndSaveDefaultLayout(ctx context.Context, userID int) (entity.UserLayout, error) {
	defaultLayout, err := m.repo.GetDefaultLayout(ctx)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("не удалось получить стандартный макет: %w", err)
	}
	defaultLayout.UserID = fmt.Sprintf("%d", userID)
	defaultLayout.LayoutID = fmt.Sprintf("default-%d", userID)
	defaultLayout.Creator = userID
	err = m.repo.UpdateLayout(ctx, defaultLayout)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("не удалось сохранить стандартный макет: %w", err)
	}
	return defaultLayout, nil
}
