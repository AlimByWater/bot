package auth

import (
	"context"
	"elysium/internal/entity"
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
