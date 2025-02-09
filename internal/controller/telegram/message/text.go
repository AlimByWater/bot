package message

import (
	"elysium/internal/usecase/use_message"
	"log/slog"
	"reflect"
	"strconv"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

type Text struct {
	logger *slog.Logger
	cache  interface {
		LoadAndDelete(key string) (value any, loaded bool)
	}
}

func NewText(
	cache interface {
		LoadAndDelete(key string) (value any, loaded bool)
	},

) *Text {
	return &Text{
		cache: cache,
	}
}

func (h *Text) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h *Text) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		handler, loaded := h.cache.LoadAndDelete(
			strconv.Itoa(int(update.Message.Chat.ID)) + ":" +
				strconv.Itoa(int(update.Message.From.ID)),
		)
		if !loaded {
			message := telegoutil.Message(
				update.Message.Chat.ChatID(),
				use_message.GL.TextPlaceholder(update.Message.From.LanguageCode),
			)
			_, err := bot.SendMessage(message)
			if err != nil {
				h.logger.Error("send message", slog.String("err", err.Error()))
			}
			return
		}
		switch handler.(type) {
		case telegohandler.Handler:
			handler.(telegohandler.Handler)(bot, update)
		default:
			h.logger.Error("not a handler",
				slog.Int64("user", update.Message.From.ID),
			)
		}
	}
}

func (h *Text) Predicate() telegohandler.Predicate {
	return telegohandler.AnyMessageWithText()
}
