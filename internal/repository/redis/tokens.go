package redis

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
)

var tokenKeyPrefix = "token_user_id"

func (m *Module) SetToken(ctx context.Context, token entity.Token) error {
	txf := func(tx *redis.Tx) error {
		_, err := tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			tokenJson, err := token.MarshalJSON()
			if err != nil {
				m.logger.Error("Failed to marshal token", slog.String("error", err.Error()), slog.String("method", "SetToken"))
				return fmt.Errorf("failed to marshal token: %w", err)
			}
			pipe.HSet(ctx, fmt.Sprintf("%s:%d", tokenKeyPrefix, token.UserID), "data", tokenJson)
			return nil
		})
		return err
	}

	var err error
	for i := 0; i < maxRetries; i++ {
		err = m.client.Watch(ctx, txf, fmt.Sprintf("%s:%d", tokenKeyPrefix, token.UserID))
		if err == nil {
			return nil
		}

		if err != redis.TxFailedErr {
			continue
		}

		return fmt.Errorf("failed to save or update token: %w", err)
	}

	return err
}

func (m *Module) GetToken(ctx context.Context, userID int) (entity.Token, error) {
	if userID == 0 {
		m.logger.Error("User ID is required", slog.String("method", "GetToken"))
		return entity.Token{}, fmt.Errorf("user id is required")
	}

	tokenJSON, err := m.client.HGet(ctx, fmt.Sprintf("%s:%d", tokenKeyPrefix, userID), "data").Bytes()
	if err != nil {
		if err == redis.Nil {
			m.logger.Error("Token not found", slog.Int("userID", userID), slog.String("method", "GetToken"))
			return entity.Token{}, entity.ErrCacheTokenNotFound
		}
		return entity.Token{}, fmt.Errorf("failed to get token: %w", err)
	}

	var token entity.Token
	err = token.UnmarshalJSON(tokenJSON)
	if err != nil {
		m.logger.Error("Failed to unmarshal token", slog.String("error", err.Error()), slog.String("method", "GetToken"))
		return entity.Token{}, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	token.UserID = userID

	return token, nil
}

func (m *Module) AllTokens(ctx context.Context) ([]entity.Token, error) {
	var tokens []entity.Token
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = m.client.Scan(ctx, cursor, tokenKeyPrefix+"*", 0).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		for _, key := range keys {
			tokenJSON, err := m.client.HGet(ctx, key, "data").Bytes()
			if err != nil {
				m.logger.Error("Failed to get token data", slog.String("key", key), slog.String("error", err.Error()))
				continue
			}

			var token entity.Token
			err = token.UnmarshalJSON(tokenJSON)
			if err != nil {
				m.logger.Error("Failed to unmarshal token", slog.String("key", key), slog.String("error", err.Error()))
				continue
			}

			userID, err := strconv.Atoi(key[len(tokenKeyPrefix+":"):])
			if err != nil {
				m.logger.Error("Failed to parse user ID from key", slog.String("key", key), slog.String("error", err.Error()))
				continue
			}
			token.UserID = userID

			tokens = append(tokens, token)
		}

		if cursor == 0 { // no more keys
			break
		}
	}

	return tokens, nil
}
