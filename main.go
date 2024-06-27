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
	"arimadj-helper/internal/usecase/arimadj"
	"arimadj-helper/internal/usecase/demethra"
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
	configModule := config.New(
		httpCfg,
		arimaDJCfg,
		demethraCfg,
	)

	/************ USECASE *************/
	demethraUC := demethra.New(demethraCfg)
	arimaDJUC := arimadj.New(arimaDJCfg)

	/************ CONTROLLER *************/
	httpModule := http.New(httpCfg,
		api.NewGroup(nil,
			api_methods.NewSubmitMethod(demethraUC),
		),
	)

	/************ APP *************/

	app := application.New(loggerModule, configModule, envModule)
	app.AddUsecases(
		arimaDJUC,
		demethraUC,
	)

	// добавляем контроллеры
	app.AddControllers(
		httpModule,
	)

	app.Run()
}
