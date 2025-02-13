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
	progress_manager "elysium/internal/controller/telegram/emoji-gen-utils/progress-manager"
	"elysium/internal/controller/telegram/group"
	"elysium/internal/controller/telegram/inline_keyboard"
	message_handlers "elysium/internal/controller/telegram/message"
	"elysium/internal/controller/telegram/middleware"
	"elysium/internal/controller/telegram/payments"
	"elysium/internal/repository/clickhouse"
	"elysium/internal/repository/clickhouse/bot_updates_insert"
	"elysium/internal/repository/clickhouse/transaction_audit_insert"
	"elysium/internal/repository/postgres"
	"elysium/internal/repository/postgres/elysium"
	"elysium/internal/repository/redis"
	"elysium/internal/usecase/emoji-gen/userbot"
	"elysium/internal/usecase/services"
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
	transactionsAuditInsert := transaction_audit_insert.NewInsertTable(loggerModule)
	botUpdatesInsert := bot_updates_insert.NewInsertTable()

	clickhouseRepo := clickhouse.New(clickhouseCfg)
	clickhouseRepo.AddInsertTable(botUpdatesInsert, transactionsAuditInsert)

	elysiumRepo := elysium.NewRepository(transactionsAuditInsert)
	//sc := soundcloud.New(soundcloudCfg)

	postgresql := postgres.New(postgresCfg, elysiumRepo)
	redisCache := redis.New(redisCfg)

	/*********************************/
	/************ USECASE ************/
	/*********************************/
	servicesUC := services.New(redisCache, elysiumRepo)
	messageUC := use_message.New(messageCfg)
	userUC := users.New(redisCache, elysiumRepo, servicesUC)

	useCache := use_cache.New()

	progressManager := progress_manager.NewManager()

	// emoji-gen modules
	userBot := userbot.NewBot(elysiumRepo, userBotCfg)
	processingUC := processing.NewProcessingModule(loggerModule)

	/*********************************/
	/********** CONTROLLER ***********/
	/*********************************/
	// mainStart := command.NewStart(messageUC)
	emojiDmStart := command.NewStartEmojiDM()
	buyTokens := payments.NewBuyTokens(userUC, useCache)
	emojiMsgTracker := group.NewEmojiMessageTracker(userBot, useCache)
	cancelEmojiHandler := inline_keyboard.NewCancelEmojiPackGeneration(progressManager)
	myPacks := inline_keyboard.NewMyEmojiPacks(useCache, elysiumRepo)
	emojiPackDelete := inline_keyboard.NewEmojiPackDelete(elysiumRepo)

	emojiDM := group.NewEmojiDM(useCache, userUC, servicesUC, processingUC, elysiumRepo, userBot, progressManager)
	emojiChat := group.NewEmojiChat(useCache, userUC, processingUC, elysiumRepo, userBot, progressManager)

	saveUpdateMiddleware := middleware.NewSaveUpdate(botUpdatesInsert)
	saveUserMiddleware := middleware.NewSaveUser(useCache, userUC)

	preCheckout := payments.NewPreCheckout(userUC)
	successPayment := payments.NewSuccessPayment(userUC, useCache)

	textMessageHandler := message_handlers.NewText(useCache)
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
			myPacks,
			emojiPackDelete,

			emojiDM,
			emojiChat,
			emojiMsgTracker,
			// commStart,
			emojiDmStart,
			cancelEmojiHandler,

			preCheckout,
			successPayment,

			textMessageHandler,
		},
	)
	/*********************************/
	/********** APPLICATION **********/
	/*********************************/
	app := application.New(loggerModule, configModule, envModule)

	// Добавление репозиториев и usecase
	app.AddRepositories(
		clickhouseRepo,
		postgresql,
		redisCache,
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
