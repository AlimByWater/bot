package application

import (
	"context"
	"log/slog"
	"os/signal"
	"reflect"
	"syscall"
)

const moduleKey = "module"
const actionKey = "action"

type controller interface {
	Init(context.Context, context.CancelFunc, *slog.Logger) error
	Run()
	Shutdown() error
}

type repository interface {
	Init(context.Context, *slog.Logger) error
	Close() error
}

type usecase interface {
	Init(context.Context, *slog.Logger) error
}

type env interface {
	Init() (interface{ Config() ([]byte, error) }, error)
}

type config interface {
	Init(interface{ Config() ([]byte, error) }) error
}

// Application основная структура приложения.
// Контролирует инициализацию, запуск и завершение всех модулей.
type Application struct {
	controllers  []controller
	repositories []repository
	usecases     []usecase
	ctx          context.Context
	stop         context.CancelFunc
	logger       *slog.Logger
	config       config
	env          env
}

// New конструктор
func New(l *slog.Logger, cfg config, e env) *Application {
	return &Application{logger: l, config: cfg, env: e}
}

// AddControllers добавляет контроллеры
func (app *Application) AddControllers(controllers ...controller) {
	app.controllers = controllers
}

// AddRepositories добавляет репозитории
func (app *Application) AddRepositories(repositories ...repository) {
	app.repositories = repositories
}

// AddUsecases добавляет модули юзкейсов
func (app *Application) AddUsecases(usecases ...usecase) {
	app.usecases = usecases
}

// Run запускает приложение
func (app *Application) Run() {
	app.logger.Info("Application running", slog.String(actionKey, "run"))

	defer app.shutdown()

	app.ctx, app.stop = signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	// Инициализация переменных окружения
	storage, err := app.env.Init()
	if err != nil {
		app.logger.Error(err.Error(), slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.env)).Type().PkgPath()))
		return
	}
	app.logger.Debug("Env init", slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.env)).Type().PkgPath()))

	// Инициализация конфигов
	err = app.config.Init(storage)
	if err != nil {
		app.logger.Error(err.Error(), slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.config)).Type().PkgPath()))
		return
	}
	app.logger.Debug("Config init", slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.config)).Type().PkgPath()))

	// Инициализация модулей репозиториев
	for i := range app.repositories {
		err = app.repositories[i].Init(app.ctx, app.logger)
		if err != nil {
			app.logger.Error(err.Error(), slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.repositories[i])).Type().PkgPath()))
			return
		}
		app.logger.Debug("Repository init", slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.repositories[i])).Type().PkgPath()))
	}

	// Инициализация модулей юзкейсов
	for i := range app.usecases {
		err = app.usecases[i].Init(app.ctx, app.logger)
		if err != nil {
			app.logger.Error(err.Error(), slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.usecases[i])).Type().PkgPath()))
			return
		}
		app.logger.Debug("Usecase init", slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.usecases[i])).Type().PkgPath()))
	}

	// Инициализация модулей контроллеров
	for i := range app.controllers {
		err = app.controllers[i].Init(app.ctx, app.stop, app.logger)
		if err != nil {
			app.logger.Error(err.Error(), slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.controllers[i])).Type().PkgPath()))
			return
		}
		app.logger.Debug("Controller init", slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.controllers[i])).Type().PkgPath()))
	}

	// Запуска контроллеров
	for i := range app.controllers {
		app.controllers[i].Run()
		app.logger.Debug("Controller run", slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.controllers[i])).Type().PkgPath()))
	}

	// Блокирование функции. Ожидаем сигнала о завершении приложения
	<-app.ctx.Done()

	return
}

func (app *Application) shutdown() {
	app.logger.Info("Application shutdown", slog.String(actionKey, "shutdown"))

	app.stop()

	for i := range app.controllers {
		err := app.controllers[i].Shutdown()
		if err != nil {
			app.logger.Error(err.Error(), slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.controllers[i])).Type().PkgPath()))
		}
		app.logger.Debug("Controller shutdown", slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.controllers[i])).Type().PkgPath()))
	}

	for i := range app.repositories {
		err := app.repositories[i].Close()
		if err != nil {
			app.logger.Error(err.Error(), slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.repositories[i])).Type().PkgPath()))
		}
		app.logger.Debug("Repository close", slog.String(moduleKey, reflect.Indirect(reflect.ValueOf(app.repositories[i])).Type().PkgPath()))
	}
}
