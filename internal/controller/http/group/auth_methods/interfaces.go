package auth_methods

import (
	"context"
	"elysium/internal/entity"
)

type authUC interface {
	GenerateTokenForTelegram(ctx context.Context, telegramLogin entity.TelegramLoginInfo) (entity.Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (entity.Token, error)
}
