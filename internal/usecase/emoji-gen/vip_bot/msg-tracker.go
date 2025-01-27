package vip_bot

import (
	"context"
	"github.com/go-telegram/bot/models"
	"strings"
	"sync"
	"time"
)

// функционал с msgTracker нужен для того чтобы обойти ограничения телеграма и tdLib
// когда мы отправляем сообщение с эмоджи-композицией (SendMessageWithEmojisToBot) мы не можем получить id этого сообщения
// поэтому мы сохраняем его id в карту msgTracker для последующего получения и пересылки пользователю

var msgTracker sync.Map

func (d *DBot) TrackEmojiMessage(update *models.Update) {
	msgID := update.Message.ID
	url := ""

	for _, e := range update.Message.Entities {
		if e.Type == models.MessageEntityTypeTextLink {
			if strings.HasPrefix(e.URL, "https://t.me/addemoji/") {
				url = e.URL
				break
			}
		}
	}

	msgTracker.Store(url, msgID)
}

func (d *DBot) waitForEmojiMessageAndForwardIt(ctx context.Context, sendTo int64, packLink string) bool {
	t := time.Now()
	for {
		if time.Since(t) > time.Second*10 {
			return false
		}
		msgIDRaw, ok := msgTracker.LoadAndDelete("https://t.me/addemoji/" + packLink)
		if !ok {
			time.Sleep(time.Second)
			continue
		}

		msgID, ok := msgIDRaw.(int)
		if !ok {
			time.Sleep(time.Second)
			continue
		}

		d.forwardMessage(ctx, d.userBot.GetID(), sendTo, msgID)
		return true
	}
}
