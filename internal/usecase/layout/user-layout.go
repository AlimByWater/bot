package layout

import (
	"arimadj-helper/internal/entity"
	"context"
	"errors"
)

var ErrNoPermission = errors.New("you don't have permission to edit this layout")

func (m *Module) GetUserLayout(ctx context.Context, userID, initiatorUserID int) (entity.UserLayout, error) {
	// Реализация получения макета пользователя
	return entity.UserLayout{}, nil
}

func (m *Module) UpdateLayoutFull(ctx context.Context, userID, initiatorUserID int, updatedLayout entity.UserLayout) error {
	currentLayout, err := m.GetUserLayout(ctx, userID, initiatorUserID)
	if err != nil {
		return err
	}

	if !m.hasEditPermission(currentLayout, initiatorUserID) {
		return ErrNoPermission
	}

	// Сохраняем создателя и редакторов
	updatedLayout.Creator = currentLayout.Creator
	updatedLayout.Editors = currentLayout.Editors

	// Здесь должна быть логика сохранения обновленного макета
	// Например, сохранение в базу данных

	return nil
}

func (m *Module) AddEditor(ctx context.Context, layoutID string, editorID int) error {
	// Реализация добавления редактора
	return nil
}

func (m *Module) RemoveEditor(ctx context.Context, layoutID string, editorID int) error {
	// Реализация удаления редактора
	return nil
}

func (m *Module) IsEditor(ctx context.Context, layoutID string, userID int) (bool, error) {
	// Реализация проверки, является ли пользователь редактором
	return false, nil
}

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
