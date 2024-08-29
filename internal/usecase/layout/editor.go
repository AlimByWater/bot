package layout

import (
	"context"
	"errors"
	"fmt"
)

// AddLayoutEditor добавляет редактора к макету
func (m *Module) AddLayoutEditor(ctx context.Context, layoutID string, initiatorUserID, editorID int) error {
	// Проверяем, имеет ли инициатор права на редактирование макета
	layout, err := m.GetLayout(ctx, layoutID, initiatorUserID)
	if err != nil {
		return fmt.Errorf("failed to get layout: %w", err)
	}

	// Проверяем, не является ли editorID уже редактором
	for _, editor := range layout.Editors {
		if editor == editorID {
			return nil
		}
	}

	// Добавляем нового редактора
	layout.Editors = append(layout.Editors, editorID)

	// Сохраняем обновленный макет
	err = m.repo.UpdateLayout(ctx, layout)
	if err != nil {
		return fmt.Errorf("failed to update layout: %w", err)
	}

	return nil
}

// RemoveLayoutEditor удаляет редактора из макета
func (m *Module) RemoveLayoutEditor(ctx context.Context, layoutID string, initiatorUserID, editorID int) error {
	// Проверяем, имеет ли инициатор права на редактирование макета
	layout, err := m.GetLayout(ctx, layoutID, initiatorUserID)
	if err != nil {
		return fmt.Errorf("failed to get layout: %w", err)
	}

	// Ищем редактора для удаления
	found := false
	for i, editor := range layout.Editors {
		if editor == editorID {
			// Удаляем редактора из списка
			layout.Editors = append(layout.Editors[:i], layout.Editors[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return errors.New("editor not found in this layout")
	}

	// Сохраняем обновленный макет
	err = m.repo.UpdateLayout(ctx, layout)
	if err != nil {
		return fmt.Errorf("failed to update layout: %w", err)
	}

	return nil
}
