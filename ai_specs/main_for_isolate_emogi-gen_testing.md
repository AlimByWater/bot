# High-Level Objective
- Создать handler для создания пака, на основе логики из з файла `internal/usecase/emoji-gen/vip_bot/cmd-emoji.go`.

# Implementation Notes
- Перенеси все нужные интерфейсы из `internal/usecase/emoji-gen/vip_bot/bot.go` в структуру и конструктор EmojiDM.
- Используйте аналогичный стиль и подход, как в других хэндлерах (например, `internal/controller/telegram/command/start-driptech.go`).
- Для взаимодействия с телеграм апи используй "github.com/mymmrac/telego"
# Low-level tasks
> ordered from start to finish  
> generate edits one at a time, so we don't overlap any changes


1. Добавь необходимые интерфейсы и методы (в EmojiDM) использующиеся в `HandleEmojiCommandForDM`. Пример методов которые нужно перенести: prepareWorkingEnvironment, handleNewPack, sendMessageByBot.   

2. Реализуй логику из HandleEmojiCommandForDM в EmojiDM.Handler используя библиотеку "github.com/mymmrac/telego"
```go
    selectedEmojis := d.processor.GenerateEmojiMessage(emojiMetaRows, stickerSet, emojiArgs)
    _, err := d.userBot.SendMessageWithEmojisToBot(ctx, strconv.FormatInt(d.tgbotApi.Self.ID, 10), emojiArgs.Width, emojiArgs.PackLink, selectedEmojis)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            return
        }
        d.logger.Error("Failed to send message with emojis",
            slog.String("err", err.Error()),
            slog.String("pack_link", emojiArgs.PackLink),
            slog.Int64("user_id", emojiArgs.TelegramUserID))
    } else {
        sent := d.waitForEmojiMessageAndForwardIt(ctx, update.Message.From.ID, emojiArgs.PackLink)
        if sent {
            return
        }
    }
	

	d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, fmt.Sprintf("Ваш пак\n%s", "https://t.me/addemoji/"+emojiArgs.PackLink), nil)
```
 