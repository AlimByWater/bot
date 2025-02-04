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
	"elysium/internal/controller/telegram/emoji-gen-utils/processing"
	"elysium/internal/controller/telegram/group"
	"elysium/internal/controller/telegram/middleware"
	"elysium/internal/repository/clickhouse"
	"elysium/internal/repository/clickhouse/bot_updates_insert"
	"elysium/internal/repository/postgres"
	"elysium/internal/repository/postgres/elysium"
	"elysium/internal/repository/redis"
	"elysium/internal/usecase/emoji-gen/userbot"
	"elysium/internal/usecase/use_cache"
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
	clickhouseCfg := config_module.NewClickhouseConfig()
	userBotCfg := config_module.NewUserBotConfig()

	configModule := config.New(
		driptechCfg,
		messageCfg,
		postgresCfg,
		redisCfg,
		clickhouseCfg,
		userBotCfg,
	)

	/*********************************/
	/********** REPOSITORY ***********/
	/*********************************/
	elysiumRepo := elysium.NewRepository()
	//sc := soundcloud.New(soundcloudCfg)

	postgresql := postgres.New(postgresCfg, elysiumRepo)
	redisCache := redis.New(redisCfg)

	clickhouseRepo := clickhouse.New(clickhouseCfg)

	botUpdatesInsert := bot_updates_insert.NewInsertTable()
	clickhouseRepo.AddInsertTable(botUpdatesInsert)

	/*********************************/
	/************ USECASE ************/
	/*********************************/
	messageUC := use_message.New(messageCfg)
	userUC := users.New(redisCache, elysiumRepo)

	useCache := use_cache.New()

	// emoji-gen modules
	userBot := userbot.NewBot(elysiumRepo, userBotCfg)
	processingUC := processing.NewProcessingModule(loggerModule)

	/*********************************/
	/********** CONTROLLER ***********/
	/*********************************/
	commStart := command.NewStart(messageUC)
	buyTokens := command.NewBuyTokens(messageUC)
	emojiMsgTracker := group.NewEmojiMessageTracker(userBot, useCache)

	emojiDM := group.NewEmojiDM(useCache, messageUC, userUC, processingUC, elysiumRepo, userBot)

	saveUpdateMiddleware := middleware.NewSaveUpdate(botUpdatesInsert)
	saveUserMiddleware := middleware.NewSaveUser(userUC)

	driptechBot := telegram.New(
		driptechCfg,
		[]telegram.Middleware{
			saveUpdateMiddleware,
			saveUserMiddleware,
		},
		[]telegram.Command{
			buyTokens,
		},
		[]telegram.GroupHandle{},
		[]telegram.Handle{
			emojiDM,
			emojiMsgTracker,
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
		clickhouseRepo,
	)
	app.AddUsecases(
		userUC,
		messageUC,
		useCache,
		userBot,
	)
	app.AddControllers(
		driptechBot,
	)

	// Запуск приложения
	app.Run()
}
