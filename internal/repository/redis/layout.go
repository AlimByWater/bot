package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"arimadj-helper/internal/entity"
)

// GetLayout получает макет из кэша по ID
func (m *Module) GetLayout(ctx context.Context, layoutID string) (interface{}, error) {
	val, err := m.client.Get(ctx, fmt.Sprintf("layout:%s", layoutID)).Bytes()
	if err != nil {
		return nil, err
	}

	var layout entity.UserLayout
	decoder := gob.NewDecoder(bytes.NewReader(val))
	err = decoder.Decode(&layout)
	if err != nil {
		return nil, err
	}

	return layout, nil
}

// SetLayout сохраняет макет в кэш
func (m *Module) SetLayout(ctx context.Context, layoutID string, value interface{}, expiration time.Duration) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(value)
	if err != nil {
		return err
	}

	return m.client.Set(ctx, fmt.Sprintf("layout:%s", layoutID), buf.Bytes(), expiration).Err()
}

// DeleteLayout удаляет макет из кэша
func (m *Module) DeleteLayout(ctx context.Context, layoutID string) error {
	return m.client.Del(ctx, fmt.Sprintf("layout:%s", layoutID)).Err()
}
func init() {
	gob.Register(entity.UserLayout{})
	// Зарегистрируйте здесь другие типы, если это необходимо
}
