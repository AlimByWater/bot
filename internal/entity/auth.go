package entity

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

var (
	ErrExpiredRefreshToken = fmt.Errorf("refresh token is expired")
	ErrInvalidToken        = fmt.Errorf("invalid token")
)

type Token struct {
	UserID       int    `json:"user_id"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type CustomClaims struct {
	TokenType TokenType `json:"type"`
	jwt.RegisteredClaims
}

type TelegramLoginInfo struct {
	TelegramID int64  `json:"telegram_id"`
	InitData   string `json:"init_data"`
}

type TelegramRefreshTokenInfo struct {
	RefreshToken string `json:"refresh_token"`
}
