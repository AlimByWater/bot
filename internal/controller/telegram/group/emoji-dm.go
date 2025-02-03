package group

import (
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"log/slog"
	"reflect"
)

type EmojiDM struct {
	logger  *slog.Logger
	message interface {
		Error(langCode string) (msg string)
		BalanceBtn(langCode string) (msg string)
		BotsListBtn(langCode string) (msg string)
		SupportBtn(langCode string) (msg string)
		BuyTokensBtn(langCode string) (msg string)
		StartDripTech(langCode string) (msg string)
	}
}

func NewEmojiDM(
	message interface {
		Error(langCode string) (msg string)
		BalanceBtn(langCode string) (msg string)
		BotsListBtn(langCode string) (msg string)
		SupportBtn(langCode string) (msg string)
		BuyTokensBtn(langCode string) (msg string)
		StartDripTech(langCode string) (msg string)
	}) *EmojiDM {
	return &EmojiDM{
		message: message,
	}
}

func (h *EmojiDM) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h *EmojiDM) Command() string     { return "emoji" }
func (h *EmojiDM) Description() string { return "Start DripTech Bot" }

func (h *EmojiDM) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		//var lang string
		//var chat telego.Chat

	}
}

func (h *EmojiDM) Predicate() telegohandler.Predicate {
	return telegohandler.And(privateChatPredicate(), telegohandler.CommandEqual(h.Command()))

}

func privateChatPredicate() telegohandler.Predicate {
	return func(update telego.Update) bool {
		if update.Message == nil {
			return false
		}

		return update.Message.Chat.Type == "private"
	}
}
