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

func (m *Module) UpdateLayoutFull(ctx context.Context, userID int, layout entity.UserLayout) error {
	// Реализация обновления макета
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
