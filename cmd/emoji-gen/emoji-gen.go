package main

import (
	"elysium/internal/application"
	"elysium/internal/application/config"
	"elysium/internal/application/config/config_module"
	"elysium/internal/application/env"
	"elysium/internal/application/env/local"
	"elysium/internal/application/logger"
	"elysium/internal/repository/postgres"
	"elysium/internal/repository/postgres/elysium"
	"elysium/internal/repository/redis"
	"elysium/internal/usecase/telegram"
	"elysium/internal/usecase/users"
	"log/slog"
	"os"
)

func main() {
	// Инициализация логирования
	loggerModule := logger.New(
		logger.Options{
			AppName: "emoji-gen",
			Writer:  os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)

	// Инициализация окружения и конфигурации
	localEnv := local.New()
	envModule := env.New(localEnv)

	telegramCfg := config_module.NewTelegramConfig()
	postgresCfg := config_module.NewPostgresConfig()
	redisCfg := config_module.NewRedisConfig()

	configModule := config.New(
		telegramCfg,
		postgresCfg,
		redisCfg,
	)

	// Инициализация репозиториев
	elysiumRepo := elysium.NewRepository()
	postgresModule := postgres.New(postgresCfg, elysiumRepo)
	redisCache := redis.New(redisCfg)

	// Инициализация usecase
	usersUC := users.New(redisCache, elysiumRepo)
	telegramUC := telegram.NewManager(telegramCfg, elysiumRepo, usersUC)

	// Создание приложения
	app := application.New(loggerModule, configModule, envModule)

	// Добавление репозиториев и usecase
	app.AddRepositories(postgresModule, redisCache)
	app.AddUsecases(
		usersUC,
		telegramUC,
	)

	// Запуск приложения
	app.Run()
}
