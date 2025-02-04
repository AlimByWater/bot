package group

import (
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"log/slog"
	"reflect"
	"strings"
)

// EmojiMessageTracker - нужен что-бы принимать сообщения от userBot с эмоджи и сохранять их MessageID в кэш, для дальнейшей пересылки пользователю
type EmojiMessageTracker struct {
	logger          *slog.Logger
	userBotIDcached int64
	userBot         interface {
		GetID() int64
	}
	cache interface {
		Store(key string, value any)
	}
}

func NewEmojiMessageTracker(
	userBot interface {
		GetID() int64
	},
	cache interface {
		Store(key string, value any)
	},
) *EmojiMessageTracker {
	return &EmojiMessageTracker{
		userBot: userBot,
		cache:   cache,
	}
}

func (h EmojiMessageTracker) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h EmojiMessageTracker) Command() string     { return "start" }
func (h EmojiMessageTracker) Description() string { return "Start DripTech Bot" }

func (h EmojiMessageTracker) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		msgID := update.Message.MessageID
		for _, e := range update.Message.Entities {
			if e.Type == telego.EntityTypeTextLink {
				if strings.HasPrefix(e.URL, "https://t.me/addemoji/") {
					h.cache.Store(e.URL, msgID)
				}
			}
		}

	}
}

func (h EmojiMessageTracker) Predicate() telegohandler.Predicate {
	return h.predicateForUserBotID()
}

func (h EmojiMessageTracker) predicateForUserBotID() telegohandler.Predicate {
	return func(update telego.Update) bool {
		if update.Message == nil {
			return false
		}

		if h.userBotIDcached != 0 && h.userBotIDcached == update.Message.From.ID {
			return true
		}

		if update.Message.From.ID == h.userBot.GetID() {
			h.userBotIDcached = h.userBot.GetID()
			return true
		}

		return false
	}
}
