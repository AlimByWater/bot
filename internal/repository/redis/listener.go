package redis

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

var ErrTelegramIDRequired = fmt.Errorf("telegram id is required")
var ErrListenerNotFound = fmt.Errorf("listener not found")
var ErrIncrimentReachedMaxNumber = fmt.Errorf("increment reached max number of retries")

const (
	maxRetries = 3
)

// SaveOrUpdateListener сохраняет или обновляет listener
// если telegramID не указан, возвращает ErrTelegramIDRequired
// если initTimestamp не указан, устанавливает текущее время, если lastActivity не указан, устанавливает текущее время
// таким образом можно использовать эту функцию для обновления времени последней активности
func (m *Module) SaveOrUpdateListener(ctx context.Context, c entity.ListenerCache) error {
	if c.TelegramID == 0 {
		return ErrTelegramIDRequired
	}

	// сначала проверяем, есть ли уже такой listener
	// если нет, то создаем новый
	// если есть, то обновляем last_activity
	txf := func(tx *redis.Tx) error {
		_, err := tx.HGet(ctx, fmt.Sprintf("listener:%d", c.TelegramID), "last_activity").Result()
		if err != nil && err != redis.Nil {
			return fmt.Errorf("failed to get listener: %w", err)
		}

		// > если listener не найден, то создаем новый
		if err == redis.Nil {
			_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				if c.Payload.InitTimestamp == 0 {
					c.Payload.InitTimestamp = time.Now().Unix()
				}

				if c.Payload.LastActivity == 0 {
					c.Payload.LastActivity = time.Now().Unix()
				}

				_, err := pipe.HSet(ctx, fmt.Sprintf("listener:%d", c.TelegramID), c.Payload).Result()
				if err != nil {
					return fmt.Errorf("pipline: failed to save listener: %w", err)
				}
				return nil
			})
			return err
		}

		// если найден - обновляем last_activity
		_, err = tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			if c.Payload.LastActivity == 0 {
				c.Payload.LastActivity = time.Now().Unix()
			}

			pipe.HSet(ctx, fmt.Sprintf("listener:%d", c.TelegramID), "last_activity", c.Payload.LastActivity)
			return nil
		})
		return err
	}

	for i := 0; i < maxRetries; i++ {
		err := m.client.Watch(ctx, txf, fmt.Sprintf("listener:%d", c.TelegramID))
		if err == nil {
			return nil
		}

		if err != redis.TxFailedErr {
			continue
		}

		return fmt.Errorf("failed to save or update listener: %w", err)
	}

	return ErrIncrimentReachedMaxNumber
}

func (m *Module) GetListenerByTelegramID(ctx context.Context, telegramID int64) (entity.ListenerCache, error) {
	if telegramID == 0 {
		return entity.ListenerCache{}, ErrTelegramIDRequired
	}
	var payload entity.ListenerCachePayload
	err := m.client.HGetAll(ctx, fmt.Sprintf("listener:%d", telegramID)).Scan(&payload)
	if err != nil {
		return entity.ListenerCache{}, fmt.Errorf("failed to get listener by telegram id: %w", err)
	}

	return entity.ListenerCache{
		TelegramID: telegramID,
		Payload:    payload,
	}, nil
}

func (m *Module) GetListenerLastActivityByTelegramID(ctx context.Context, telegramID int64) (int64, error) {
	if telegramID == 0 {
		return 0, ErrTelegramIDRequired
	}
	res, err := m.client.HGet(ctx, fmt.Sprintf("listener:%d", telegramID), "last_activity").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get last activity by telegram id: %w", ErrListenerNotFound)
	}

	lastActivity, err := strconv.ParseInt(res, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse last activity: %w", err)

	}

	return lastActivity, nil
}

func (m *Module) GetListenersCount(ctx context.Context) (int64, error) {
	count := 0
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = m.client.Scan(ctx, cursor, "listener:*", 0).Result()
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

func (m *Module) RemoveListenerTelegramID(ctx context.Context, telegramID int64) error {
	if telegramID == 0 {
		return ErrTelegramIDRequired
	}

	_, err := m.client.Del(ctx, fmt.Sprintf("listener:%d", telegramID)).Result()
	if err != nil {
		return fmt.Errorf("failed to remove listener by telegram id: %w", err)
	}

	return nil
}

// GetAllCurrentListeners возвращает всех текущих слушателей
// достаточно трудоемкая операция, так как нужно просканировать все ключи и получить их значения
// но пока слушателей не десятки тысяч - думаю норм (я ничего об это не знаю)
func (m *Module) GetAllCurrentListeners(ctx context.Context) ([]entity.ListenerCache, error) {
	var listeners []entity.ListenerCache
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = m.client.Scan(ctx, cursor, "listener:*", 0).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		for _, key := range keys {
			var payload entity.ListenerCachePayload
			err := m.client.HGetAll(ctx, key).Scan(&payload)
			if err != nil {
				return nil, fmt.Errorf("failed to get listener by telegram id: %w", err)
			}

			telegramID, err := strconv.ParseInt(key[9:], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse telegram id: %w", err)
			}

			listeners = append(listeners, entity.ListenerCache{
				TelegramID: telegramID,
				Payload:    payload,
			})
		}

		if cursor == 0 { // no more keys
			break
		}
	}

	return listeners, nil
}
