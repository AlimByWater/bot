package services

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"log/slog"
)

type cacheUC interface {
	SaveServiceCache(ctx context.Context, service entity.Service) error
	GetServiceCache(ctx context.Context, botID int64, serviceName string) (entity.Service, error)
	RemoveServiceCache(ctx context.Context, botID int64, serviceName string) error
}

type repository interface {
	ListServicesByBotID(ctx context.Context, botID int64) ([]entity.Service, error)
	GetAllServices(ctx context.Context) ([]entity.Service, error)
	InitServices(ctx context.Context, services []entity.Service) error
}

type Module struct {
	logger *slog.Logger
	ctx    context.Context

	cache cacheUC
	repo  repository
}

func New(cache cacheUC, repo repository) *Module {
	return &Module{
		cache: cache,
		repo:  repo,
	}
}

func (m *Module) initServices(ctx context.Context) ([]entity.Service, error) {
	services := []entity.Service{
		{
			BotID:       -1007894673045,
			Name:        "emoji-generator",
			Description: "Генерация телеграм эмоджи-паков из любого контента",
			Price:       1500, //15 рублей за один пак
			IsActive:    true,
		},
	}

	err := m.repo.InitServices(ctx, services)
	if err != nil {
		return nil, fmt.Errorf("failed to init services: %w", err)
	}

	return services, nil
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.String("module", "📱 services"))

	// Загружаем все сервисы из БД в кэш при инициализации
	//services, err := m.repo.GetAllServices(ctx)
	//if err != nil {
	//	return fmt.Errorf("failed to get all services: %w", err)
	//}

	// TODO после того как определюсь с набором сервисов эту функцию нужно убрать
	services, err := m.initServices(ctx)
	if err != nil {
		return fmt.Errorf("failed to init services: %w", err)
	}

	// Сохраняем каждый сервис в кэш
	for _, service := range services {
		err = m.cache.SaveServiceCache(ctx, service)
		if err != nil {
			m.logger.Error("failed to save service to cache",
				slog.Int64("bot_id", service.BotID),
				slog.String("service_name", service.Name),
				slog.String("error", err.Error()),
			)
		}
	}

	m.logger.Info("services cache initialized",
		slog.Int("total_services", len(services)),
	)

	return nil
}

// GetService получает сервис по имени для конкретного бота
func (m *Module) GetService(ctx context.Context, botID int64, serviceName string) (entity.Service, error) {
	// Пробуем получить из кэша
	service, err := m.cache.GetServiceCache(ctx, botID, serviceName)
	if err == nil {
		return service, nil
	}

	// Если в кэше нет, получаем все сервисы из БД
	services, err := m.repo.ListServicesByBotID(ctx, botID)
	if err != nil {
		return entity.Service{}, err
	}

	// Ищем нужный сервис и сохраняем все в кэш
	for _, service := range services {
		// Сохраняем каждый сервис в кэш
		err = m.cache.SaveServiceCache(ctx, service)
		if err != nil {
			m.logger.Error("failed to save service to cache",
				slog.Int64("bot_id", botID),
				slog.String("service_name", service.Name),
				slog.String("error", err.Error()),
			)
		}

		// Если это искомый сервис, запоминаем его для возврата
		if service.Name == serviceName {
			return service, nil
		}
	}

	return entity.Service{}, entity.ErrServiceNotFound
}

// GetServices получает все сервисы для бота
func (m *Module) GetServices(ctx context.Context, botID int64) ([]entity.Service, error) {
	// Получаем сервисы из БД
	services, err := m.repo.ListServicesByBotID(ctx, botID)
	if err != nil {
		return nil, err
	}

	// Сохраняем каждый сервис в кэш
	for _, service := range services {
		err = m.cache.SaveServiceCache(ctx, service)
		if err != nil {
			m.logger.Error("failed to save service to cache",
				slog.Int64("bot_id", botID),
				slog.String("service_name", service.Name),
				slog.String("error", err.Error()),
			)
		}
	}

	return services, nil
}
