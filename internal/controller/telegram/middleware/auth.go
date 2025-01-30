package middleware

import (
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"log/slog"
	"reflect"
)

type Auth struct {
	logger   *slog.Logger
	userData interface {
		SetContact(userId int64, firstName, lastName, username, langCode string) (text string, ok bool)
	}
}

func NewAuth(
	userData interface {
		SetContact(userId int64, firstName, lastName, username, langCode string) (text string, ok bool)
	},
) *Auth {
	return &Auth{
		userData: userData,
	}
}

func (h *Auth) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h *Auth) Handler() telegohandler.Middleware {
	return func(bot *telego.Bot, update telego.Update, next telegohandler.Handler) {
		var (
			id        int64
			firstName string
			lastName  string
			username  string
			langCode  string
			chatId    telego.ChatID
		)
		if update.Message != nil {
			id = update.Message.From.ID
			firstName = update.Message.From.FirstName
			lastName = update.Message.From.LastName
			username = update.Message.From.Username
			langCode = update.Message.From.LanguageCode
			chatId = update.Message.Chat.ChatID()
		} else if update.CallbackQuery != nil {
			id = update.CallbackQuery.From.ID
			firstName = update.CallbackQuery.From.FirstName
			lastName = update.CallbackQuery.From.LastName
			username = update.CallbackQuery.From.Username
			langCode = update.CallbackQuery.From.LanguageCode
			chat := update.CallbackQuery.Message.GetChat()
			chatId = chat.ChatID()
		} else {
			return
		}

		text, ok := h.userData.SetContact(
			id,
			firstName,
			lastName,
			username,
			langCode,
		)

		if !ok {
			_, err := bot.SendMessage(
				telegoutil.Message(
					chatId,
					text,
				),
			)
			if err != nil {
				h.logger.Error("send message", slog.String("err", err.Error()))
			}
			return
		}
		next(bot, update)
	}
}
