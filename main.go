package main

import (
	"elysium/internal/application"
	"elysium/internal/application/config"
	"elysium/internal/application/config/config_module"
	"elysium/internal/application/env"
	"elysium/internal/application/env/local"
	"elysium/internal/application/logger"
	"elysium/internal/controller/http"
	api "elysium/internal/controller/http/group"
	"elysium/internal/controller/http/group/auth_methods"
	"elysium/internal/controller/http/group/layout_methods"
	"elysium/internal/controller/http/group/song_methods"
	"elysium/internal/controller/http/group/tampermonkey_methods"
	web_app_methods "elysium/internal/controller/http/group/web-app_methods"
	"elysium/internal/controller/scheduler"
	"elysium/internal/controller/scheduler/scheduler_job"
	"elysium/internal/repository/downloader"
	"elysium/internal/repository/postgres"
	"elysium/internal/repository/postgres/elysium"
	"elysium/internal/repository/redis"
	"elysium/internal/usecase/auth"
	"elysium/internal/usecase/demethra"
	"elysium/internal/usecase/layout"
	"elysium/internal/usecase/users"
	"log/slog"
	"os"
)

func main() {
	loggerModule := logger.New(
		logger.Options{
			AppName: "bot-manager",
			Writer:  os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)

	/************ CONFIG *************/
	localEnv := local.New()
	envModule := env.New(
		localEnv,
	)

	httpCfg := config_module.NewHttpConfig()
	arimaDJCfg := config_module.NewArimaDJConfig()
	demethraCfg := config_module.NewDemethraConfig()
	postgresCfg := config_module.NewPostgresConfig()
	soundcloudCfg := config_module.NewSoundcloudConfig()
	redisCfg := config_module.NewRedisConfig()
	authCfg := config_module.NewAuthConfig()
	downloaderCfg := config_module.NewDownloaderConfig()
	configModule := config.New(
		arimaDJCfg,
		httpCfg,
		demethraCfg,
		postgresCfg,
		soundcloudCfg,
		redisCfg,
		authCfg,
		downloaderCfg,
	)

	/************ REPOSITORY *************/
	elysiumRepo := elysium.NewRepository()
	//sc := soundcloud.New(soundcloudCfg)
	downloaderGrpc := downloader.New(downloaderCfg)

	postgresql := postgres.New(postgresCfg, elysiumRepo)
	redisCache := redis.New(redisCfg)

	/************ USECASE *************/
	//arimaDJUC := arimadj.New(arimaDJCfg)
	layoutUC := layout.New(redisCache, elysiumRepo)
	usersUC := users.New(redisCache, elysiumRepo, layoutUC)
	authUC := auth.NewModule(authCfg, redisCache, elysiumRepo, usersUC)
	demethraUC := demethra.New(demethraCfg, elysiumRepo, redisCache, downloaderGrpc, usersUC)

	/************ CONTROLLER *************/
	httpModule := http.New(httpCfg,
		api.NewSongGroup(authUC,
			song_methods.NewSongByURL(demethraUC), // /api/song/url/:url
		),
		api.NewLayoutGroup(authUC,
			layout_methods.NewLayoutByID(layoutUC),         // /api/layout/:id
			layout_methods.NewLayoutByName(layoutUC),       // /api/layout/name/:name
			layout_methods.NewUpdateLayout(layoutUC),       // /api/layout/:id
			layout_methods.NewAddLayoutEditor(layoutUC),    // /api/layout/:id/editor
			layout_methods.NewRemoveLayoutEditor(layoutUC), // /api/layout/:id/editor/:editorId
			layout_methods.NewGetUserLayout(layoutUC),      // /api/layout/user/:userID
		),
		api.NewAuthGroup(authUC,
			auth_methods.NewGenerateMethod(authUC),
			auth_methods.NewRefreshMethod(authUC),
		),
		api.NewWebAppGroup(authUC,
			web_app_methods.NewWebAppState(demethraUC),
			web_app_methods.NewWebAppEvent(demethraUC),
			web_app_methods.NewWebsocketEvent(usersUC, demethraUC),
		),
		api.NewGroup(authCfg,
			tampermonkey_methods.NewSubmitMethod(demethraUC),
		),
		//api.NewUserGroup(authUC,
		//	user_method.NewGetUser(usersUC),
		//),
	)

	activeListenerCheckJob := scheduler_job.NewListenerCheckJob(demethraUC)

	schedularModule := scheduler.New(nil,
		activeListenerCheckJob,
	)

	/************ APP *************/

	app := application.New(loggerModule, configModule, envModule)

	app.AddRepositories(postgresql, redisCache, downloaderGrpc)

	app.AddUsecases(
		layoutUC,
		//arimaDJUC,
		demethraUC,
		authUC,
		usersUC,
	)

	// добавляем контроллеры
	app.AddControllers(
		httpModule,
		schedularModule,
	)

	app.Run()
}
