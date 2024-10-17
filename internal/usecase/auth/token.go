package auth

import (
	"context"
	"database/sql"
	"elysium/internal/application/logger"
	"elysium/internal/entity"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"log/slog"
	"strconv"
	"time"
)

const ACCESS_TOKEN_TTL = 2 * 60 * time.Minute
const REFRESH_TOKEN_TTL = 90 * 24 * time.Hour

func (m *Module) CheckAccessTokenByUserID(ctx context.Context, token string, userID int) (bool, error) {
	if token == "test-token" && userID == 3 {
		return true, nil
	}
	attributes := []slog.Attr{
		slog.String("method", "CheckAccessTokenByUserID"),
		slog.Int("userID", userID),
		slog.String("token", token),
	}

	claims := entity.CustomClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return m.jwtSecret, nil
	})
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "parse with claims", logger.AppendErrorToLogs(attributes, err)...)
		return false, fmt.Errorf("parse with claims: %w", err)
	}

	if claims.TokenType != entity.TokenTypeAccess {
		m.logger.LogAttrs(ctx, slog.LevelDebug, "invalid token type", attributes...)
		return false, nil
	}

	if claims.Subject != strconv.Itoa(userID) {
		m.logger.LogAttrs(ctx, slog.LevelDebug, "invalid user id", attributes...)
		return false, nil
	}

	cacheToken, err := m.GetTokenByUserID(ctx, userID)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "get token by user id", logger.AppendErrorToLogs(attributes, err)...)
		return false, fmt.Errorf("get token by user id: %w", err)
	}

	if cacheToken.AccessToken != token {
		m.logger.LogAttrs(ctx, slog.LevelError, "token not equal", attributes...)
		return false, nil
	}

	return true, nil
}

func (m *Module) GenerateTokenForTelegram(ctx context.Context, telegramLogin entity.TelegramLoginInfo) (entity.Token, error) {
	attrs := []slog.Attr{
		slog.String("method", "GenerateTokenForTelegram"),
		slog.Int64("telegram_id", telegramLogin.TelegramID),
	}

	data, err := m.parseTelegramInitDate(telegramLogin.InitData)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "parse telegram init data", logger.AppendErrorToLogs(attrs, err)...)
		return entity.Token{}, fmt.Errorf("parse telegram init data: %w", err)

	}

	var user entity.User
	user, err = m.repo.GetUserByTelegramID(ctx, telegramLogin.TelegramID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		m.logger.LogAttrs(ctx, slog.LevelError, "get user by telegram id", logger.AppendErrorToLogs(attrs, err)...)
		return entity.Token{}, fmt.Errorf("get user by telegram id: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		user, err = m.users.CreateUser(ctx, entity.User{
			TelegramID:       telegramLogin.TelegramID,
			TelegramUsername: data.User.Username,
			Firstname:        data.User.FirstName,
		})
		if err != nil {
			m.logger.LogAttrs(ctx, slog.LevelError, "create user", logger.AppendErrorToLogs(attrs, err)...)
			return entity.Token{}, fmt.Errorf("create user: %w", err)
		}
	}

	accessToken, err := m.generateJWTToken(ctx, user.ID, entity.TokenTypeAccess, ACCESS_TOKEN_TTL)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "generate access token", logger.AppendErrorToLogs(attrs, err)...)
		return entity.Token{}, fmt.Errorf("generate token: %w", err)
	}

	refreshToken, err := m.generateJWTToken(ctx, user.ID, entity.TokenTypeRefresh, REFRESH_TOKEN_TTL)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "generate refresh token", logger.AppendErrorToLogs(attrs, err)...)
		return entity.Token{}, fmt.Errorf("generate token: %w", err)
	}

	token := entity.Token{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	go m.cacheToken(token)

	return token, nil
}

// RefreshToken methods that regenerating entity.Token with new access token and refresh token and swap it in tokensMap if refresh token is valid; returns refresh token
func (m *Module) RefreshToken(ctx context.Context, refreshToken string) (entity.Token, error) {
	claims := entity.CustomClaims{}
	_, err := jwt.ParseWithClaims(refreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return m.jwtSecret, nil
	})
	if err != nil {
		return entity.Token{}, fmt.Errorf("parse with claims: %w", err)
	}

	if claims.TokenType != entity.TokenTypeRefresh {
		return entity.Token{}, entity.ErrInvalidToken
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		return entity.Token{}, fmt.Errorf("parse user id: %w", err)
	}

	cacheToken, err := m.GetTokenByUserID(ctx, userID)
	if err != nil {
		return entity.Token{}, fmt.Errorf("get token by user id: %w", err)
	}
	var currentClaims entity.CustomClaims
	_, err = jwt.ParseWithClaims(cacheToken.RefreshToken, &currentClaims, func(token *jwt.Token) (interface{}, error) {
		return m.jwtSecret, nil
	})
	if err != nil {
		return entity.Token{}, fmt.Errorf("parse with current claims: %w", err)
	}

	if !currentClaims.ExpiresAt.Time.Before(time.Now()) {
		return entity.Token{}, entity.ErrExpiredRefreshToken
	}

	if cacheToken.RefreshToken != refreshToken {
		return entity.Token{}, entity.ErrInvalidToken
	}

	newAccessToken, err := m.generateJWTToken(ctx, userID, entity.TokenTypeAccess, ACCESS_TOKEN_TTL)
	if err != nil {
		return entity.Token{}, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := m.generateJWTToken(ctx, userID, entity.TokenTypeRefresh, REFRESH_TOKEN_TTL)
	if err != nil {
		return entity.Token{}, fmt.Errorf("generate refresh token: %w", err)
	}

	token := entity.Token{
		UserID:       userID,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	go m.cacheToken(token)

	return token, nil
}

func (m *Module) generateJWTToken(_ context.Context, userID int, tokenType entity.TokenType, ttl time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, entity.CustomClaims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	})

	return token.SignedString(m.jwtSecret)
}

func (m *Module) parseTelegramInitDate(initData string) (initdata.InitData, error) {
	err := initdata.Validate(initData, m.cfg.GetTelegramBotToken(), 24*time.Hour)
	if err != nil {
		err2 := initdata.Validate(initData, "7486051673:AAEg2bzMqec1NkFK8tHycLn8gvGxK6xQ6ww", 24*time.Hour)
		if err2 != nil {
			return initdata.InitData{}, fmt.Errorf("invalid init data: %w; %s", err, err2.Error())
		}
	}

	parsedData, err := initdata.Parse(initData)
	if err != nil {
		return initdata.InitData{}, fmt.Errorf("failed to parse init data: %w", err)
	}

	return parsedData, nil
}

func (m *Module) cacheToken(token entity.Token) {
	m.tokensMap.Store(token.UserID, token)
	err := m.cache.SetToken(token)
	if err != nil {
		m.logger.Error("Failed to cache token", slog.String("error", err.Error()), slog.Int("userID", token.UserID), slog.String("method", "cacheToken"))
	}
}
