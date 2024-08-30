package redis

import (
	"context"
	"fmt"
	"time"

	"arimadj-helper/internal/entity"
)

// GetLayout получает макет из кэша по ID
func (m *Module) GetLayout(ctx context.Context, layoutID string) (interface{}, error) {
	key := fmt.Sprintf("layout:%s", layoutID)
	values, err := m.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("layout not found")
	}

	layout := entity.UserLayout{
		ID:          values["id"],
		UserID:      values["user_id"],
		Name:        values["name"],
		Description: values["description"],
		// Добавьте здесь другие поля структуры UserLayout
	}

	return layout, nil
}

// SetLayout сохраняет макет в кэш
func (m *Module) SetLayout(ctx context.Context, layoutID string, value interface{}, expiration time.Duration) error {
	layout, ok := value.(entity.UserLayout)
	if !ok {
		return fmt.Errorf("invalid layout type")
	}

	key := fmt.Sprintf("layout:%s", layoutID)
	_, err := m.client.HSet(ctx, key,
		"id", layout.ID,
		"user_id", layout.UserID,
		"name", layout.Name,
		"description", layout.Description,
		// Добавьте здесь другие поля структуры UserLayout
	).Result()
	if err != nil {
		return err
	}

	return m.client.Expire(ctx, key, expiration).Err()
}

// DeleteLayout удаляет макет из кэша
func (m *Module) DeleteLayout(ctx context.Context, layoutID string) error {
	return m.client.Del(ctx, fmt.Sprintf("layout:%s", layoutID)).Err()
}
