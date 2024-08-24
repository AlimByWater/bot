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
	"arimadj-helper/internal/controller/http/group/tampermonkey_methods"
	"arimadj-helper/internal/controller/http/group/web-app_methods"
	"arimadj-helper/internal/controller/scheduler"
	"arimadj-helper/internal/controller/scheduler/scheduler_job"
	"arimadj-helper/internal/repository/postgres"
	"arimadj-helper/internal/repository/postgres/elysium"
	"arimadj-helper/internal/repository/redis"
	"arimadj-helper/internal/repository/soundcloud"
	"arimadj-helper/internal/usecase/auth"
	"arimadj-helper/internal/usecase/demethra"
	"arimadj-helper/internal/usecase/users"
	"log/slog"
	"os"
)

//func maintest() {
//	loggerModule := logger.New(
//		logger.Options{
//			AppName: "bot-manager",
//			Writer:  os.Stdout,
//			HandlerOptions: &slog.HandlerOptions{
//				Level: slog.LevelDebug,
//			},
//		},
//	)
//
//	/************ CONFIG *************/
//	localEnv := local.New()
//	envModule := env.New(
//		localEnv,
//	)
//
//	httpCfg := config_module.NewHttpConfig()
//	arimaDJCfg := config_module.NewArimaDJConfig()
//	demethraCfg := config_module.NewDemethraConfig()
//	postgresCfg := config_module.NewPostgresConfig()
//	configModule := config.New(
//		httpCfg,
//		arimaDJCfg,
//		demethraCfg,
//		postgresCfg,
//	)
//
//	/************ POSTGRES *************/
//
//	/************ USECASE *************/
//	demethraUC := demethra.New(demethraCfg)
//	arimaDJUC := arimadj.New(arimaDJCfg)
//
//	/************ CONTROLLER *************/
//	httpModule := http.New(httpCfg,
//		api.NewGroup(nil,
//			tampermonkey_methods.NewSubmitMethod(demethraUC),
//		),
//	)
//
//	/************ APP *************/
//
//	app := application.New(loggerModule, configModule, envModule)
//	app.AddUsecases(
//		arimaDJUC,
//		demethraUC,
//	)
//
//	// добавляем контроллеры
//	app.AddControllers(
//		httpModule,
//	)
//
//	app.Run()
//}

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
	demethraCfg := config_module.NewDemethraConfig()
	postgresCfg := config_module.NewPostgresConfig()
	soundcloudCfg := config_module.NewSoundcloudConfig()
	redisCfg := config_module.NewRedisConfig()
	authCfg := config_module.NewAuthConfig()
	configModule := config.New(
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
	demethraUC := demethra.New(demethraCfg, elysiumRepo, redisCache, sc)
	authUC := auth.NewModule(authCfg, elysiumRepo)
	usersUC := users.New(redisCache)

	/************ CONTROLLER *************/
	httpModule := http.New(httpCfg,
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
	)

	activeListenerCheckJob := scheduler_job.NewListenerCheckJob(demethraUC)

	schedularModule := scheduler.New(nil,
		activeListenerCheckJob,
	)

	/************ APP *************/

	app := application.New(loggerModule, configModule, envModule)

	app.AddRepositories(postgresql, redisCache, sc)

	app.AddUsecases(
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
