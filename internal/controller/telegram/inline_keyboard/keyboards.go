package inline_keyboard

import (
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

func GetEmojiBotStartKeyboard(
	langCode string,
	message interface {
		BuyTokensBtn(lang string) string
		CreatePackkInfoBtn(lang string) string
		MyPacksBtn(lang string) string
	},
) *telego.InlineKeyboardMarkup {
	return telegoutil.InlineKeyboard(
		telegoutil.InlineKeyboardRow(
			telegoutil.InlineKeyboardButton(message.BuyTokensBtn(langCode)).WithURL("t.me/driptechbot?start=start"), // Купить токены - должен ссылаться на @driptechbot
			telegoutil.InlineKeyboardButton("FAQ").WithCallbackData("faq"),                                          // FAQ // TODO: не реализовано
		),
		telegoutil.InlineKeyboardRow(
			telegoutil.InlineKeyboardButton(message.CreatePackkInfoBtn(langCode)).WithCallbackData("info"), // Создать пак // TODO: не реализовано
			telegoutil.InlineKeyboardButton(message.MyPacksBtn(langCode)).WithCallbackData("my_packs"),     // Мои паки
		),
	)
}
