package middleware

import (
	"elysium/internal/entity"
	"encoding/json"
	"fmt"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"log/slog"
	"reflect"
	"strconv"
	"time"
)

type SaveUpdate struct {
	logger        *slog.Logger
	saveUpdaterer interface {
		SaveUpdate(botUpdate entity.BotUpdate) error
	}
}

func NewSaveUpdate(
	saveUpdaterer interface {
		SaveUpdate(botUpdate entity.BotUpdate) error
	},
) *SaveUpdate {
	return &SaveUpdate{
		saveUpdaterer: saveUpdaterer,
	}
}

func (m *SaveUpdate) AddLogger(logger *slog.Logger) {
	m.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(m)).Type().PkgPath()))
}

func (m *SaveUpdate) Handler() telegohandler.Middleware {
	return func(bot *telego.Bot, update telego.Update, next telegohandler.Handler) {
		go func() {
			updCopy, err := update.CloneSafe()
			if err != nil {
				m.logger.Error("Failed to parse bot ID",
					slog.String("error", err.Error()))
				next(bot, update)
				return
			}

			botUser, ok := updCopy.Context().Value(entity.BotSelfCtxKey).(*telego.User)
			if !ok || botUser == nil {
				m.logger.Error("bot info not found in context")
				return
			}

			botID, err := strconv.ParseInt(fmt.Sprintf("-100%d", botUser.ID), 10, 64)
			if err != nil {
				m.logger.Error("Failed to parse bot ID",
					slog.Int64("bot_id", botUser.ID),
					slog.String("error", err.Error()))
				return
			}

			updCopyAsBytes, err := json.Marshal(updCopy)

			botUpdate := entity.BotUpdate{
				BotID:      botID,
				UpdateTime: time.Now(),
				Payload:    string(updCopyAsBytes),
			}

			err = m.saveUpdaterer.SaveUpdate(botUpdate)
			if err != nil {
				m.logger.Error("Failed to save update",
					slog.String("error", err.Error()))
			}
		}()

		next(bot, update)
	}
}
