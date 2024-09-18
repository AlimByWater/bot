package redis

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var defaultExpiredTime = time.Hour * 24

func (m *Module) GetLayoutByUserID(ctx context.Context, userID int) (entity.UserLayout, error) {
	if userID == 0 {
		return entity.UserLayout{}, entity.ErrCacheLayoutIDRequired
	}

	layoutJSON, err := m.client.HGet(ctx, fmt.Sprintf("layout_user_id:%d", userID), "data").Result()
	if err != nil {
		if err == redis.Nil {
			return entity.UserLayout{}, entity.ErrLayoutNotFound
		}
		return entity.UserLayout{}, fmt.Errorf("failed to get layout: %w", err)
	}

	var layout entity.UserLayout
	err = layout.UnmarshalJSON([]byte(layoutJSON))
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("failed to unmarshal layout: %w", err)
	}

	return layout, nil
}

func (m *Module) GetLayoutByName(ctx context.Context, layoutName string) (entity.UserLayout, error) {
	if layoutName == "" {
		return entity.UserLayout{}, entity.ErrCacheLayoutIDRequired
	}

	layoutJSON, err := m.client.HGet(ctx, fmt.Sprintf("layout_name:%s", layoutName), "data").Result()
	if err != nil {
		if err == redis.Nil {
			return entity.UserLayout{}, entity.ErrLayoutNotFound
		}
		return entity.UserLayout{}, fmt.Errorf("failed to get layout: %w", err)
	}

	var layout entity.UserLayout
	err = layout.UnmarshalJSON([]byte(layoutJSON))
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("failed to unmarshal layout: %w", err)
	}

	return layout, nil
}

// SaveOrUpdateLayout сохраняет или обновляет макет
func (m *Module) SaveOrUpdateLayout(ctx context.Context, layout entity.UserLayout) error {
	if layout.ID == 0 {
		return entity.ErrCacheLayoutIDRequired
	}

	txf := func(tx *redis.Tx) error {
		_, err := tx.HGet(ctx, fmt.Sprintf("layout:%d", layout.ID), "user_id").Result()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get layout: %w", err)
		}

		layoutJSON, err := layout.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal layout: %w", err)
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.HSet(ctx, fmt.Sprintf("layout:%d", layout.ID), "data", string(layoutJSON))
			pipe.HSet(ctx, fmt.Sprintf("layout_name:%s", layout.Name), "data", string(layoutJSON))
			pipe.HSet(ctx, fmt.Sprintf("layout_user_id:%d", layout.Creator), "data", string(layoutJSON))

			pipe.Expire(ctx, fmt.Sprintf("layout:%d", layout.ID), defaultExpiredTime)
			pipe.Expire(ctx, fmt.Sprintf("layout_name:%s", layout.Name), defaultExpiredTime)
			pipe.Expire(ctx, fmt.Sprintf("layout_user_id:%d", layout.Creator), defaultExpiredTime)
			return nil
		})
		return err
	}

	for i := 0; i < maxRetries; i++ {
		err := m.client.Watch(ctx, txf, fmt.Sprintf("layout:%d", layout.ID))
		if err == nil {
			return nil
		}

		if err != redis.TxFailedErr {
			continue
		}

		return fmt.Errorf("failed to save or update layout: %w", err)
	}

	return entity.ErrCacheIncrementReachedMaxNumber
}

// GetLayout получает макет из кэша по ID
func (m *Module) GetLayout(ctx context.Context, layoutID int) (entity.UserLayout, error) {
	if layoutID == 0 {
		return entity.UserLayout{}, entity.ErrCacheLayoutIDRequired
	}

	layoutJSON, err := m.client.HGet(ctx, fmt.Sprintf("layout:%d", layoutID), "data").Result()
	if err != nil {
		if err == redis.Nil {
			return entity.UserLayout{}, entity.ErrLayoutNotFound
		}
		return entity.UserLayout{}, fmt.Errorf("failed to get layout: %w", err)
	}

	var layout entity.UserLayout
	err = layout.UnmarshalJSON([]byte(layoutJSON))
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("failed to unmarshal layout: %w", err)
	}

	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("failed to unmarshal layout: %w", err)
	}

	return layout, nil
}

// DeleteLayout удаляет макет из кэша
func (m *Module) DeleteLayout(ctx context.Context, layoutID int) error {
	if layoutID == 0 {
		return entity.ErrCacheLayoutIDRequired
	}

	_, err := m.client.Del(ctx, fmt.Sprintf("layout:%d", layoutID)).Result()
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
			err = layout.UnmarshalJSON([]byte(layoutJSON))
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
