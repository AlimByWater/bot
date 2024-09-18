package redis_test

import (
	"context"
	"elysium/internal/entity"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSetToken(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	token := entity.Token{
		UserID:       5,
		AccessToken:  "",
		RefreshToken: "",
	}

	err := redisModule.SetToken(token)
	require.NoError(t, err)

	cachedToken, err := redisModule.GetToken(context.Background(), token.UserID)
	require.NoError(t, err)
	require.Equal(t, token.AccessToken, cachedToken.AccessToken)
	require.Equal(t, token.RefreshToken, cachedToken.RefreshToken)
}

func TestAllTokens(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	//// Preparing a token to be retrieved
	//tokens := []entity.Token{
	//	{
	//		UserID:       rand.Int(),
	//		AccessToken:  "access-token",
	//		RefreshToken: "refresh-token",
	//	},
	//	{
	//		UserID:       rand.Int(),
	//		AccessToken:  "access-token",
	//		RefreshToken: "refresh-token",
	//	},
	//}
	//
	//for _, token := range tokens {
	//	err := redisModule.SetToken( token)
	//	require.NoError(t, err)
	//}

	// Retrieving all tokens
	retrievedTokens, err := redisModule.AllTokens(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, retrievedTokens)
	t.Log(retrievedTokens[0])
}
