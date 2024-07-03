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
	"arimadj-helper/internal/controller/http/group/api_methods"
	"arimadj-helper/internal/repository/postgres"
	"arimadj-helper/internal/repository/postgres/elysium"
	"arimadj-helper/internal/repository/soundcloud"
	"arimadj-helper/internal/usecase/demethra"
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
//			api_methods.NewSubmitMethod(demethraUC),
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
	configModule := config.New(
		httpCfg,
		demethraCfg,
		postgresCfg,
	)

	/************ REPOSITORY *************/
	elysiumRepo := elysium.NewRepository()
	sc := soundcloud.New(soundcloudCfg)

	postgresql := postgres.New(postgresCfg, elysiumRepo)

	/************ USECASE *************/
	demethraUC := demethra.New(demethraCfg, elysiumRepo, sc)

	/************ CONTROLLER *************/
	httpModule := http.New(httpCfg,
		api.NewGroup(nil,
			api_methods.NewSubmitMethod(demethraUC),
		),
	)

	/************ APP *************/

	app := application.New(loggerModule, configModule, envModule)

	app.AddRepositories(postgresql, sc)

	app.AddUsecases(
		demethraUC,
	)

	// добавляем контроллеры
	app.AddControllers(
		httpModule,
	)

	app.Run()
}
