package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"arimadj-helper/internal/entity"
)

// GetLayout получает макет из кэша по ID
func (m *Module) GetLayout(ctx context.Context, layoutID string) (interface{}, error) {
	val, err := m.client.Get(ctx, fmt.Sprintf("layout:%s", layoutID)).Result()
	if err != nil {
		return nil, err
	}

	var layout entity.UserLayout
	err = json.Unmarshal([]byte(val), &layout)
	if err != nil {
		return nil, err
	}

	return layout, nil
}

// SetLayout сохраняет макет в кэш
func (m *Module) SetLayout(ctx context.Context, layoutID string, value interface{}, expiration time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return m.client.Set(ctx, fmt.Sprintf("layout:%s", layoutID), jsonValue, expiration).Err()
}

// DeleteLayout удаляет макет из кэша
func (m *Module) DeleteLayout(ctx context.Context, layoutID string) error {
	return m.client.Del(ctx, fmt.Sprintf("layout:%s", layoutID)).Err()
}
