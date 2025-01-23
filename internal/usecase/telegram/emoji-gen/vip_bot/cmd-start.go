package vip_bot

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/inline"
	"github.com/go-telegram/ui/keyboard/reply"
	"log/slog"
)

func (d *DBot) HandleStartCommand(ctx context.Context, b *bot.Bot, update *models.Update) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Привет, *" + bot.EscapeMarkdown(update.Message.From.FirstName) + "*",
		ParseMode:   models.ParseModeMarkdown,
		ReplyMarkup: d.menuButtons(ctx),
	})

	startKeyboard := d.startKeyboard(ctx)

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "Добро пожаловать на сервер.\nЯ ⁂VIP бот, а это значит:\n ⁂ Твои запросы обрабатываются вне очереди\n ⁂ Ты можешь получать готовые эмодзи-композиции в ЛС\n ⁂ Ты можешь именовать паки без префикса (параметр name=[])\n⁂ пока что все",
		ReplyMarkup: startKeyboard,
	})

	if err != nil {
		d.logger.Error("Failed to send message to DM", slog.String("username", update.Message.From.Username), slog.Int64("user_id", update.Message.From.ID), slog.String("err", err.Error()))
	}
}

func (d *DBot) startKeyboard(ctx context.Context) models.ReplyMarkup {
	kb := inline.New(d.b.BotApi).
		Row().
		Button("Создать пак", []byte("emoji"), d.onEmojiSelect).
		Button("Мои паки", []byte("packs"), d.onRemovePacksSelect)
	//Button()

	return kb
}

func (d *DBot) menuButtons(_ context.Context) models.ReplyMarkup {
	kb := reply.New(
		reply.WithPrefix("reply_keyboard"),
		reply.IsSelective(),
		reply.IsPersistent())

	return kb.Row().
		Button("⁂ Меню", d.b.BotApi, bot.MatchTypeExact, d.onMenuSelect)

}

func (d *DBot) onMenuSelect(ctx context.Context, b *bot.Bot, update *models.Update) {
	keyboard := d.startKeyboard(ctx)

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      update.Message.Chat.ID,
		Text:        "тут текст меню",
		ReplyMarkup: keyboard,
	})

	if err != nil {
		d.logger.Error("Failed to send message to DM", slog.String("username", update.Message.From.Username), slog.Int64("user_id", update.Message.From.ID), slog.String("err", err.Error()))
	}
}
