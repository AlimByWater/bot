package users

import (
	"context"
	"elysium/internal/entity"
)

func (m *Module) CanGenerateEmojiPack(ctx context.Context, user entity.User, chatID int64) (bool, error) {
	return true, nil
}
