package main

import (
	"elysium/internal/application"
	"elysium/internal/application/config"
	"elysium/internal/application/config/config_module"
	"elysium/internal/application/env"
	"elysium/internal/application/env/local"
	"elysium/internal/application/logger"
	"elysium/internal/controller/telegram"
	"elysium/internal/controller/telegram/command"
	"elysium/internal/controller/telegram/middleware"
	"elysium/internal/repository/postgres"
	"elysium/internal/repository/postgres/elysium"
	"elysium/internal/repository/redis"
	"elysium/internal/usecase/use_message"
	"elysium/internal/usecase/users"
	"log/slog"
	"os"
)

func main() {
	// Инициализация логирования
	loggerModule := logger.New(
		logger.Options{
			AppName: "driptech-bot",
			Writer:  os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)

	// Инициализация окружения и конфигурации
	localEnv := local.New()
	envModule := env.New(localEnv)

	driptechCfg := config_module.NewDripTechBotConfig()
	messageCfg := config_module.NewMessageConfig()
	postgresCfg := config_module.NewPostgresConfig()
	redisCfg := config_module.NewRedisConfig()

	configModule := config.New(
		driptechCfg,
		messageCfg,
		postgresCfg,
		redisCfg,
	)

	/*********************************/
	/********** REPOSITORY ***********/
	/*********************************/
	elysiumRepo := elysium.NewRepository()
	//sc := soundcloud.New(soundcloudCfg)

	postgresql := postgres.New(postgresCfg, elysiumRepo)
	redisCache := redis.New(redisCfg)

	/*********************************/
	/************ USECASE ************/
	/*********************************/
	messageUC := use_message.New(messageCfg)
	userUC := users.New(redisCache, elysiumRepo)

	/*********************************/
	/********** CONTROLLER ***********/
	/*********************************/
	commStart := command.NewStart(
		messageUC,
	)

	buyTokens := command.NewBuyTokens(
		messageUC,
	)

	saveUserMiddleware := middleware.NewSaveUser(userUC)

	driptechBot := telegram.New(
		driptechCfg,
		[]telegram.Middleware{
			saveUserMiddleware,
		},
		[]telegram.Command{
			buyTokens,
		},
		[]telegram.GroupHandle{},
		[]telegram.Handle{
			commStart,
		},
	)
	/*********************************/
	/********** APPLICATION **********/
	/*********************************/
	app := application.New(loggerModule, configModule, envModule)

	// Добавление репозиториев и usecase
	app.AddRepositories(
		postgresql,
		redisCache,
	)
	app.AddUsecases(
		userUC,
		messageUC,
	)
	app.AddControllers(
		driptechBot,
	)

	// Запуск приложения
	app.Run()
}
