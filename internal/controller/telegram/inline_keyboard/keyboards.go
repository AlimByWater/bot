package inline_keyboard

import (
	"elysium/internal/usecase/use_message"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func GetEmojiBotStartKeyboard(
	langCode string,
) *telego.InlineKeyboardMarkup {
	return telegoutil.InlineKeyboard(
		telegoutil.InlineKeyboardRow(
			telegoutil.InlineKeyboardButton(use_message.GL.BuyTokensBtn(langCode)).WithURL("t.me/driptechbot?start=start"), // Купить токены - должен ссылаться на @driptechbot
			telegoutil.InlineKeyboardButton("FAQ").WithCallbackData("faq"),                                                 // FAQ // TODO: не реализовано
		),
		telegoutil.InlineKeyboardRow(
			telegoutil.InlineKeyboardButton(use_message.GL.CreatePackkInfoBtn(langCode)).WithCallbackData("info"), // Создать пак // TODO: не реализовано
			telegoutil.InlineKeyboardButton(use_message.GL.MyPacksBtn(langCode)).WithCallbackData("my_packs"),     // Мои паки
		),
	)
}
