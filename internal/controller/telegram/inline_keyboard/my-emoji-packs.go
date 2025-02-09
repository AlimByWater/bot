package inline_keyboard

import (
	"context"
	"elysium/internal/entity"
	"elysium/internal/usecase/use_message"
	"log/slog"
	"strconv"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

type MyEmojiPacks struct {
	logger *slog.Logger
	repo   interface {
		GetEmojiPacksByCreator(ctx context.Context, userID int64, deleted bool) ([]entity.EmojiPack, error)
	}
	cache interface {
		Store(key string, value any)
	}
}

func NewMyEmojiPacks(
	cache interface {
		Store(key string, value any)
	},
	repo interface {
		GetEmojiPacksByCreator(ctx context.Context, userID int64, deleted bool) ([]entity.EmojiPack, error)
	},
) *MyEmojiPacks {
	return &MyEmojiPacks{
		repo:  repo,
		cache: cache,
	}
}

func (h *MyEmojiPacks) onPackSelected() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		user := update.Message.From

		inlineKeyboard := telegoutil.InlineKeyboard(
			telegoutil.InlineKeyboardRow(
				telegoutil.InlineKeyboardButton(use_message.GL.RemovePackBtn(user.LanguageCode)).WithCallbackData("pack_delete:"+update.Message.Text),
			),
			telegoutil.InlineKeyboardRow(
				telegoutil.InlineKeyboardButton(use_message.GL.BackBtn(user.LanguageCode)).WithCallbackData(h.Command()),
			),
		)

		message := telegoutil.Message(
			telegoutil.ID(user.ID),
			use_message.GL.ChoosenPack(user.LanguageCode)+"t.me/addemoji/"+update.Message.Text,
		).WithReplyMarkup(inlineKeyboard)

		_, err := bot.SendMessage(message)
		if err != nil {
			h.logger.Error("Failed to send message", slog.String("err", err.Error()), slog.Int64("user_id", user.ID), slog.String("username", user.Username))
			return
		}
	}
}

func (h *MyEmojiPacks) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		var user *telego.User
		var messageID int
		if update.CallbackQuery != nil {
			user = &update.CallbackQuery.From
			messageID = update.CallbackQuery.Message.GetMessageID()
		} else if update.Message != nil {
			user = update.Message.From
			messageID = update.Message.GetMessageID()
		} else {
			return
		}

		_ = messageID

		packs, err := h.repo.GetEmojiPacksByCreator(update.Context(), user.ID, false)
		if err != nil {
			h.logger.Error("Failed to get emoji packs by creator", slog.String("err", err.Error()), slog.Int64("user_id", user.ID), slog.String("username", user.Username))

		}

		if len(packs) != 0 {
			// replyKeboard := telegoutil.KeyboarCols().WithSelective().WithResizeKeyboard().WithOneTimeKeyboard()
			buttons := make([]telego.KeyboardButton, 0, len(packs)+1)
			for _, pack := range packs {
				buttons = append(buttons, telego.KeyboardButton{
					Text: pack.PackLink,
				})
			}

			// buttons = append(buttons, telegoutil.KeyboardButton(use_message.GL.BackBtn(user.LanguageCode)))

			replyKeyboard := telegoutil.KeyboardGrid(telegoutil.KeyboardCols(2, buttons...)).WithOneTimeKeyboard().WithResizeKeyboard()
			message := telegoutil.Message(
				telegoutil.ID(user.ID),
				use_message.GL.ChoosePack(user.LanguageCode),
			).WithReplyMarkup(replyKeyboard)

			_, err = bot.SendMessage(message)
			if err != nil {
				h.logger.Error("Failed to send message", slog.String("err", err.Error()), slog.Int64("user_id", user.ID), slog.String("username", user.Username))
				return
			}

			h.cache.Store(strconv.FormatInt(user.ID, 10)+":"+strconv.FormatInt(user.ID, 10), h.onPackSelected())
		} else if len(packs) == 0 {

			// bot.DeleteMessage(&telego.DeleteMessageParams{
			// 	ChatID:    telegoutil.ID(user.ID),
			// 	MessageID: message.GetMessageID(),
			// })

			message := telegoutil.Message(
				telegoutil.ID(user.ID),
				use_message.GL.UserDontHavePacks(user.LanguageCode),
			).WithReplyMarkup(telegoutil.ReplyKeyboardRemove())

			_, err = bot.SendMessage(message)
			if err != nil {
				h.logger.Error("Failed to send message", slog.String("err", err.Error()), slog.Int64("user_id", user.ID), slog.String("username", user.Username))
				return
			}
		}

	}
}

func (h *MyEmojiPacks) AddLogger(logger *slog.Logger) {
	h.logger = logger
}

func (h *MyEmojiPacks) Command() string {
	return "my_packs"
}

func (h *MyEmojiPacks) Predicate() telegohandler.Predicate {
	return telegohandler.Or(
		telegohandler.CallbackDataPrefix(h.Command()),
		telegohandler.CommandEqual(h.Command()),
	)
}
