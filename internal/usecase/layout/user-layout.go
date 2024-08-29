package layout

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"time"
)

// GetUserLayout получает макет пользователя с учетом прав доступа инициатора
func (m *Module) GetUserLayout(ctx context.Context, userID, initiatorUserID int) (entity.UserLayout, error) {
	layout, err := m.repo.LayoutByUserID(ctx, userID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("не удалось получить макет пользователя: %w", err)
	}

	if !m.hasViewPermission(layout, initiatorUserID) {
		return entity.UserLayout{}, entity.ErrNoPermission
	}

	// Фильтрация элементов макета, если инициатор не является создателем или редактором
	if !m.hasEditPermission(layout, initiatorUserID) {
		filteredElements := make([]entity.LayoutElement, 0, len(layout.Layout))
		for _, element := range layout.Layout {
			if element.Public {
				filteredElements = append(filteredElements, element)
			}
		}
		layout.Layout = filteredElements
	}

	return layout, nil
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

	err = m.logLayoutChange(ctx, initiatorUserID, layoutID, "UpdateLayoutFull", fmt.Sprintf("Макет обновлен"))
	if err != nil {
		m.logger.Error("Не удалось записать изменение макета", "error", err)
	}
	return nil
}

// GetLayout получает макет по его ID с учетом прав доступа инициатора
func (m *Module) GetLayout(ctx context.Context, layoutID string, initiatorUserID int) (entity.UserLayout, error) {
	layout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("не удалось получить макет: %w", err)
	}

	if !m.hasViewPermission(layout, initiatorUserID) {
		return entity.UserLayout{}, entity.ErrNoPermission
	}

	// Фильтрация элементов макета, если инициатор не является создателем или редактором
	if !m.hasEditPermission(layout, initiatorUserID) {
		filteredElements := make([]entity.LayoutElement, 0, len(layout.Layout))
		for _, element := range layout.Layout {
			if element.Public {
				filteredElements = append(filteredElements, element)
			}
		}
		layout.Layout = filteredElements
	}

	return layout, nil
}

// AddEditor добавляет нового редактора к макету
func (m *Module) AddEditor(ctx context.Context, layoutID string, editorID int) error {
	layout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return fmt.Errorf("не удалось получить макет: %w", err)
	}

	layout.Editors = append(layout.Editors, editorID)

	err = m.repo.UpdateLayout(ctx, layout)
	if err != nil {
		return fmt.Errorf("не удалось обновить макет с новым редактором: %w", err)
	}

	err = m.logLayoutChange(ctx, layout.Creator, layoutID, "AddEditor", fmt.Sprintf("Добавлен редактор %d к макету", editorID))
	if err != nil {
		m.logger.Error("Не удалось записать изменение макета", "error", err)
	}
	return nil
}

// RemoveEditor удаляет редактора из макета
func (m *Module) RemoveEditor(ctx context.Context, layoutID string, editorID int) error {
	layout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return fmt.Errorf("не удалось получить макет: %w", err)
	}

	for i, editor := range layout.Editors {
		if editor == editorID {
			layout.Editors = append(layout.Editors[:i], layout.Editors[i+1:]...)
			break
		}
	}

	err = m.repo.UpdateLayout(ctx, layout)
	if err != nil {
		return fmt.Errorf("не удалось обновить макет после удаления редактора: %w", err)
	}

	err = m.logLayoutChange(ctx, layout.Creator, layoutID, "RemoveEditor", fmt.Sprintf("Удален редактор %d из макета", editorID))
	if err != nil {
		m.logger.Error("Не удалось записать изменение макета", "error", err)
	}
	return nil
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
