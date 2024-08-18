package auth

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
)

func (m *Module) GetTokenByUserID(_ context.Context, userID int) (entity.Token, error) {
	t, ok := m.tokensMap.Load(userID)
	if !ok {
		return entity.Token{}, fmt.Errorf("cache: get token: not found")
	}

	token, ok := t.(entity.Token)
	if !ok {
		return entity.Token{}, fmt.Errorf("cache: get token: invalid type")
	}

	return token, nil
}

//
//func (m *Module) GetTokenByTelegramID(ctx context.Context, telegramID int64) (string, error) {
//	user, err := m.repo.GetUserByTelegramID(ctx, telegramID)
//	if err != nil {
//		return "", fmt.Errorf("get user by telegram id: %w", err)
//	}
//
//	token, err := m.repo.TokenByUserID(ctx, user.ID)
//	if err != nil {
//		return "", fmt.Errorf("token by user id: %w", err)
//	}
//
//	return token.AccessToken, nil
//}
//
//func (m *Module) GenerateTokenByTelegramID(ctx context.Context, telegramID int64) (entity.Token, error) {
//	user, err := m.repo.GetUserByTelegramID(ctx, telegramID)
//	if err != nil {
//		return entity.Token{}, fmt.Errorf("get user by telegram id: %w", err)
//	}
//
//	accessToken, err := m.generateJWTToken(ctx, user.ID, entity.TokenTypeAccess, ACCESS_TOKEN_TTL)
//	if err != nil {
//		return entity.Token{}, fmt.Errorf("generate token: %w", err)
//	}
//
//	refreshToken, err := m.generateJWTToken(ctx, user.ID, entity.TokenTypeRefresh, REFRESH_TOKEN_TTL)
//	if err != nil {
//		return entity.Token{}, fmt.Errorf("generate token: %w", err)
//	}
//	return entity.Token{
//		UserID:       user.ID,
//		AccessToken:  accessToken,
//		RefreshToken: refreshToken,
//	}, nil
//}
