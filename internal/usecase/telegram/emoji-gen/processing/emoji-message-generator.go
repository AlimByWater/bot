package processing

import (
	"elysium/internal/entity"
	"github.com/go-telegram/bot/models"
)

func (m *Module) GenerateEmojiMessage(emojiMetaRows [][]entity.EmojiMeta, stickerSet *models.StickerSet, emojiArgs *entity.EmojiCommand) []entity.EmojiMeta {
	transparentCount := 0
	newEmojis := make([]entity.EmojiMeta, 0, entity.MaxStickerInMessage)
	for _, row := range emojiMetaRows {
		for _, emoji := range row {
			newEmojis = append(newEmojis, emoji)
			if emoji.Transparent {
				transparentCount++
			}
		}
	}

	// Выбираем нужные эмодзи
	selectedEmojis := make([]entity.EmojiMeta, 0, entity.MaxStickerInMessage)
	if emojiArgs.NewSet {
		selectedEmojis = newEmojis
	} else {
		// Выбираем последние 100 эмодзи из пака
		startIndex := len(stickerSet.Stickers) - entity.MaxStickerInMessage
		if startIndex < 0 {
			startIndex = 0
		}
		for _, sticker := range stickerSet.Stickers[startIndex:] {
			selectedEmojis = append(selectedEmojis, entity.EmojiMeta{
				FileID:      sticker.FileID,
				DocumentID:  sticker.CustomEmojiID,
				Transparent: false,
			})
		}
	}

	return selectedEmojis
}
