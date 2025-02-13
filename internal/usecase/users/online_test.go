package users_test

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetOnlineUsersCount(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		time.Sleep(time.Second * 15)
		onlineUsersCount := module.GetOnlineUsersCount()
		require.NotEmpty(t, onlineUsersCount)
		t.Log(onlineUsersCount)
	})
}
