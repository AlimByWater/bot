package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"arimadj-helper/internal/entity"
)

var ErrLayoutIDRequired = fmt.Errorf("layout id is required")
var ErrLayoutNotFound = fmt.Errorf("layout not found")
var ErrIncrementReachedMaxNumber = fmt.Errorf("increment reached max number of retries")

const (
	maxRetries = 3
)

// SaveOrUpdateLayout сохраняет или обновляет макет
func (m *Module) SaveOrUpdateLayout(ctx context.Context, layout entity.UserLayout) error {
	if layout.LayoutID == "" {
		return ErrLayoutIDRequired
	}

	txf := func(tx *redis.Tx) error {
		_, err := tx.HGet(ctx, fmt.Sprintf("layout:%s", layout.LayoutID), "user_id").Result()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get layout: %w", err)
		}

		layoutJSON, err := json.Marshal(layout)
		if err != nil {
			return fmt.Errorf("failed to marshal layout: %w", err)
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(ctx, fmt.Sprintf("layout:%s", layout.LayoutID), "data", string(layoutJSON))
			return nil
		})
		return err
	}

	for i := 0; i < maxRetries; i++ {
		err := m.client.Watch(ctx, txf, fmt.Sprintf("layout:%s", layout.LayoutID))
		if err == nil {
			return nil
		}

		if err != redis.TxFailedErr {
			continue
		}

		return fmt.Errorf("failed to save or update layout: %w", err)
	}

	return ErrIncrementReachedMaxNumber
}

// GetLayout получает макет из кэша по ID
func (m *Module) GetLayout(ctx context.Context, layoutID string) (entity.UserLayout, error) {
	if layoutID == "" {
		return entity.UserLayout{}, ErrLayoutIDRequired
	}

	layoutJSON, err := m.client.HGet(ctx, fmt.Sprintf("layout:%s", layoutID), "data").Result()
	if err != nil {
		if err == redis.Nil {
			return entity.UserLayout{}, ErrLayoutNotFound
		}
		return entity.UserLayout{}, fmt.Errorf("failed to get layout: %w", err)
	}

	var layout entity.UserLayout
	err = json.Unmarshal([]byte(layoutJSON), &layout)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("failed to unmarshal layout: %w", err)
	}

	return layout, nil
}

// DeleteLayout удаляет макет из кэша
func (m *Module) DeleteLayout(ctx context.Context, layoutID string) error {
	if layoutID == "" {
		return ErrLayoutIDRequired
	}

	_, err := m.client.Del(ctx, fmt.Sprintf("layout:%s", layoutID)).Result()
	if err != nil {
		return fmt.Errorf("failed to delete layout: %w", err)
	}

	return nil
}

// GetLayoutsCount возвращает количество макетов в кэше
func (m *Module) GetLayoutsCount(ctx context.Context) (int64, error) {
	count := 0
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = m.client.Scan(ctx, cursor, "layout:*", 0).Result()
		if err != nil {
			return 0, fmt.Errorf("failed to scan keys: %w", err)
		}

		count += len(keys)

		if cursor == 0 { // no more keys
			break
		}
	}

	return int64(count), nil
}

// GetAllLayouts возвращает все макеты из кэша
func (m *Module) GetAllLayouts(ctx context.Context) ([]entity.UserLayout, error) {
	var layouts []entity.UserLayout
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = m.client.Scan(ctx, cursor, "layout:*", 0).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		for _, key := range keys {
			layoutJSON, err := m.client.HGet(ctx, key, "data").Result()
			if err != nil {
				return nil, fmt.Errorf("failed to get layout data: %w", err)
			}

			var layout entity.UserLayout
			err = json.Unmarshal([]byte(layoutJSON), &layout)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal layout: %w", err)
			}

			layouts = append(layouts, layout)
		}

		if cursor == 0 { // no more keys
			break
		}
	}

	return layouts, nil
}
