package telegram

import (
	"context"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"log/slog"
)

type config interface {
	GetToken() string
	GetName() string
}

type Middleware interface {
	AddLogger(logger *slog.Logger)
	Handler() telegohandler.Middleware
}

type Command interface {
	Command() string
	Description() string
	Handle
}

type Handle interface {
	AddLogger(logger *slog.Logger)
	Handler() telegohandler.Handler
	Predicate() telegohandler.Predicate
}

type GroupHandle interface {
	AddLogger(logger *slog.Logger)
	Handler(botHandler *telegohandler.BotHandler)
}

func New(cfg config, md []Middleware, cm []Command, gh []GroupHandle, h []Handle) *Module {
	return &Module{
		cfg:         cfg,
		middleware:  md,
		commands:    cm,
		groupHandle: gh,
		handles:     h,
	}
}

type Module struct {
	ctx    context.Context
	stop   context.CancelFunc
	logger *slog.Logger
	cfg    config

	bot         *telego.Bot
	updates     <-chan telego.Update
	botHandler  *telegohandler.BotHandler
	middleware  []Middleware
	commands    []Command
	groupHandle []GroupHandle
	handles     []Handle
}

func (m *Module) Init(ctx context.Context, stop context.CancelFunc, logger *slog.Logger) (err error) {
	m.ctx = ctx
	m.stop = stop
	m.logger = logger.With(slog.String("module", "💬 EMOJI GEN VIP"))

	m.bot, err = telego.NewBot(m.cfg.GetToken(), telego.WithLogger(telegoLogger{m.logger}))
	if err != nil {
		return
	}

	m.updates, err = m.bot.UpdatesViaLongPolling(nil)
	if err != nil {
		return
	}

	m.botHandler, err = telegohandler.NewBotHandler(m.bot, m.updates)
	if err != nil {
		return
	}

	m.botHandler.Use(telegohandler.PanicRecoveryHandler(telegoRecovery{m.logger}.Handler))

	for _, mid := range m.middleware {
		mid.AddLogger(m.logger)
		m.botHandler.Use(mid.Handler())
	}

	//if m.cfg.GetName() != "" && m.cfg.GetName() != me.FirstName {
	//	err = m.bot.SetMyName(&telego.SetMyNameParams{
	//		Name: m.cfg.GetName(),
	//	})
	//	if err != nil {
	//		return
	//	}
	//}

	//if m.cfg.GetDescription() != "" {
	//	err = m.bot.SetMyDescription(&telego.SetMyDescriptionParams{
	//		Description: m.cfg.GetDescription(),
	//	})
	//	if err != nil {
	//		return
	//	}
	//}

	//if m.cfg.GetShortDescription() != "" {
	//	err = m.bot.SetMyShortDescription(&telego.SetMyShortDescriptionParams{
	//		ShortDescription: m.cfg.GetShortDescription(),
	//	})
	//	if err != nil {
	//		return
	//	}
	//}

	if len(m.commands) > 0 {
		commands := make([]telego.BotCommand, 0, len(m.commands))
		for _, c := range m.commands {
			c.AddLogger(m.logger)
			m.botHandler.Handle(c.Handler(), c.Predicate())
			commands = append(commands, telego.BotCommand{
				Command:     c.Command(),
				Description: c.Description(),
			})
		}
		err = m.bot.SetMyCommands(&telego.SetMyCommandsParams{
			Commands: commands,
		})
		if err != nil {
			return
		}
		err = m.bot.SetChatMenuButton(&telego.SetChatMenuButtonParams{
			MenuButton: &telego.MenuButtonCommands{Type: "commands"},
		})
		if err != nil {
			return
		}
	}

	for _, gh := range m.groupHandle {
		gh.AddLogger(m.logger)
		gh.Handler(m.botHandler)
	}

	for _, h := range m.handles {
		h.AddLogger(m.logger)
		m.botHandler.Handle(h.Handler(), h.Predicate())
	}

	return
}

func (m *Module) Run() {
	go m.run()
}

func (m *Module) run() {
	m.botHandler.Start()
	m.stop()
}

func (m *Module) Shutdown() (err error) {
	if m.bot != nil {
		m.bot.StopLongPolling()
	}
	if m.botHandler != nil {
		m.botHandler.Stop()
	}
	return
}
