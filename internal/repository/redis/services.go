package redis

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"github.com/redis/go-redis/v9"
)

const servicesCachePrefix = "service:"

func (m *Module) SaveServiceCache(ctx context.Context, service entity.Service) error {
	if service.BotID == 0 {
		return fmt.Errorf("botID is zero")
	}

	key := fmt.Sprintf("%s%d:%s", servicesCachePrefix, service.BotID, service.Name)

	data, err := service.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal service: %w", err)
	}

	err = m.client.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save service to cache: %w", err)
	}

	return nil
}

func (m *Module) GetServiceCache(ctx context.Context, botID int64, serviceName string) (entity.Service, error) {
	if botID == 0 {
		return entity.Service{}, fmt.Errorf("botID is zero")
	}

	key := fmt.Sprintf("%s%d:%s", servicesCachePrefix, botID, serviceName)

	data, err := m.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return entity.Service{}, entity.ErrServiceNotFound
	} else if err != nil {
		return entity.Service{}, fmt.Errorf("failed to get service from cache: %w", err)
	}

	var service entity.Service
	err = service.UnmarshalJSON([]byte(data))
	if err != nil {
		return entity.Service{}, fmt.Errorf("failed to unmarshal service data: %w", err)
	}

	return service, nil
}

func (m *Module) RemoveServiceCache(ctx context.Context, botID int64, serviceName string) error {
	if botID == 0 {
		return fmt.Errorf("botID is zero")
	}

	key := fmt.Sprintf("%s%d:%s", servicesCachePrefix, botID, serviceName)

	err := m.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove service from cache: %w", err)
	}

	return nil
}
