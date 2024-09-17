package main

import (
	"arimadj-helper/internal/application"
	"arimadj-helper/internal/application/config"
	"arimadj-helper/internal/application/config/config_module"
	"arimadj-helper/internal/application/env"
	"arimadj-helper/internal/application/env/local"
	"arimadj-helper/internal/application/logger"
	"arimadj-helper/internal/controller/http"
	api "arimadj-helper/internal/controller/http/group"
	"arimadj-helper/internal/controller/http/group/auth_methods"
	"arimadj-helper/internal/controller/http/group/layout_methods"
	"arimadj-helper/internal/controller/http/group/song_methods"
	"arimadj-helper/internal/controller/http/group/tampermonkey_methods"
	"arimadj-helper/internal/controller/http/group/user_method"
	web_app_methods "arimadj-helper/internal/controller/http/group/web-app_methods"
	"arimadj-helper/internal/controller/scheduler"
	"arimadj-helper/internal/controller/scheduler/scheduler_job"
	"arimadj-helper/internal/repository/postgres"
	"arimadj-helper/internal/repository/postgres/elysium"
	"arimadj-helper/internal/repository/redis"
	"arimadj-helper/internal/repository/soundcloud"
	"arimadj-helper/internal/usecase/arimadj"
	"arimadj-helper/internal/usecase/auth"
	"arimadj-helper/internal/usecase/demethra"
	"arimadj-helper/internal/usecase/layout"
	"arimadj-helper/internal/usecase/users"
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
	configModule := config.New(
		arimaDJCfg,
		httpCfg,
		demethraCfg,
		postgresCfg,
		soundcloudCfg,
		redisCfg,
		authCfg,
	)

	/************ REPOSITORY *************/
	elysiumRepo := elysium.NewRepository()
	sc := soundcloud.New(soundcloudCfg)

	postgresql := postgres.New(postgresCfg, elysiumRepo)
	redisCache := redis.New(redisCfg)

	/************ USECASE *************/
	arimaDJUC := arimadj.New(arimaDJCfg)
	demethraUC := demethra.New(demethraCfg, elysiumRepo, redisCache, sc)
	layoutUC := layout.New(redisCache, elysiumRepo)
	usersUC := users.New(redisCache, elysiumRepo, layoutUC)
	authUC := auth.NewModule(authCfg, redisCache, elysiumRepo, usersUC)

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
		api.NewUserGroup(authUC,
			user_method.NewGetUserLayout(layoutUC),
		),
	)

	activeListenerCheckJob := scheduler_job.NewListenerCheckJob(demethraUC)

	schedularModule := scheduler.New(nil,
		activeListenerCheckJob,
	)

	/************ APP *************/

	app := application.New(loggerModule, configModule, envModule)

	app.AddRepositories(postgresql, redisCache, sc)

	app.AddUsecases(
		layoutUC,
		arimaDJUC,
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
