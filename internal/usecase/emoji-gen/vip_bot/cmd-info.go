package vip_bot

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
)

func (d *DBot) HandleInfoCommand(ctx context.Context, update *models.Update) {
	d.sendInfoMessage(ctx, update.Message.Chat.ID, update.Message.ID)
}

func (d *DBot) sendInfoMessage(ctx context.Context, chatID int64, replyTo int) {
	infoText := `🤖 Бот для создания эмодзи-паков из картинок/видео/GIF

Отправьте медиафайл с командой /emoji и опциональными параметрами в формате param=[value]:

Параметры:
• width=[N] или w=[N] - ширина нарезки (по умолчанию 8). Чем меньше ширина, тем крупнее эмодзи
• background=[цвет] или b=[цвет] - цвет фона, который будет вырезан из изображения. Поддерживаются:
  - HEX формат: b=[0x00FF00]
  - Названия: b=[black], b=[white], b=[pink], b=[green]
• b_sim=[число] - порог схожести цвета с фоном (0-1, по умолчанию 0.1)
• b_blend=[число] - использовать смешивание цветов для удаления фона (0-1, по умолчанию 0.1)
• link=[ссылка] или l=[ссылка] - добавить эмодзи в существующий пак (должен быть создан вами)
• iphone=[false] или i=[false] - убирает оптимизация размера эмодзи под iPhone; по-умолчанию true `

	params := &bot.SendMessageParams{
		ChatID: chatID,
		Text:   infoText,
	}

	if replyTo != 0 {
		params.ReplyMarkup = &models.ReplyParameters{
			MessageID: replyTo,
			ChatID:    chatID,
		}
	}

	_, err := d.b.BotApi.SendMessage(ctx, params)
	if err != nil {
		d.logger.Error("Failed to send info message", slog.String("err", err.Error()), slog.Int64("user_id", chatID))
	}
}
