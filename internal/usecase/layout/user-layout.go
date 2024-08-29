package layout

import (
	"arimadj-helper/internal/entity"
	"context"
	"errors"
)

func (m *Module) GetUserLayout(ctx context.Context, userID int) (entity.UserLayout, error) {
	// Реализация получения макета пользователя
	return entity.UserLayout{}, nil
}

func (m *Module) UpdateLayoutFull(ctx context.Context, userID, initiatorUserID int, updatedLayout entity.UserLayout) error {
	currentLayout, err := m.GetUserLayout(ctx, userID, initiatorUserID)
	if err != nil {
		return err
	}

	isCreator := currentLayout.Creator == initiatorUserID
	isEditor := false
	for _, editor := range currentLayout.Editors {
		if editor == initiatorUserID {
			isEditor = true
			break
		}
	}

	if !isCreator && !isEditor {
		return errors.New("you don't have permission to edit this layout")
	}

	// Сохраняем создателя и редакторов
	updatedLayout.Creator = currentLayout.Creator
	updatedLayout.Editors = currentLayout.Editors

	// Здесь должна быть логика сохранения обновленного макета
	// Например, сохранение в базу данных

	return nil
}

func (m *Module) AddEditor(ctx context.Context, layoutID string, editorID string) error {
	// Реализация добавления редактора
	return nil
}

func (m *Module) RemoveEditor(ctx context.Context, layoutID string, editorID string) error {
	// Реализация удаления редактора
	return nil
}

func (m *Module) IsEditor(ctx context.Context, layoutID string, userID string) (bool, error) {
	// Реализация проверки, является ли пользователь редактором
	return false, nil
}
