package auth_methods

import (
	"arimadj-helper/internal/entity"
	"context"
)

type authUC interface {
	GenerateTokenForTelegram(ctx context.Context, telegramLogin entity.TelegramLoginInfo) (entity.Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (entity.Token, error)
}
