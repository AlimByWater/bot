package users

import (
	"context"
	"elysium/internal/entity"
)

func (m *Module) CanGenerateEmojiPack(ctx context.Context, user entity.User) (bool, error) {
	return true, nil
}
