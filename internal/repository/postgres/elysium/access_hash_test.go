package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateOrUpdateAccessHash(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	t.Run("create new access hash", func(t *testing.T) {
		ah := entity.AccessHash{
			ChatID:   "-1001111111111",
			Username: "testbot1",
			Hash:     111111111,
			PeerID:   1111111111,
		}

		t.Cleanup(func() {
			err := elysiumRepo.DeleteAccessHash(context.Background(), ah.ChatID)
			require.NoError(t, err)
		})

		err := elysiumRepo.CreateOrUpdateAccessHash(context.Background(), ah)
		require.NoError(t, err)

		saved, err := elysiumRepo.GetAccessHash(context.Background(), ah.ChatID)
		require.NoError(t, err)
		assert.Equal(t, ah.ChatID, saved.ChatID)
		assert.Equal(t, ah.Username, saved.Username)
		assert.Equal(t, ah.Hash, saved.Hash)
		assert.Equal(t, ah.PeerID, saved.PeerID)
		assert.False(t, saved.CreatedAt.IsZero())
	})

	t.Run("update existing access hash", func(t *testing.T) {
		ah := entity.AccessHash{
			ChatID:   "-1002222222222",
			Username: "testbot2",
			Hash:     222222222,
			PeerID:   2222222222,
		}

		t.Cleanup(func() {
			err := elysiumRepo.DeleteAccessHash(context.Background(), ah.ChatID)
			require.NoError(t, err)
		})

		// Создаем начальную запись
		err := elysiumRepo.CreateOrUpdateAccessHash(context.Background(), ah)
		require.NoError(t, err)

		// Обновляем запись
		ah.Username = "updatedbot2"
		ah.Hash = 222222223
		ah.PeerID = 2222222223

		err = elysiumRepo.CreateOrUpdateAccessHash(context.Background(), ah)
		require.NoError(t, err)

		saved, err := elysiumRepo.GetAccessHash(context.Background(), ah.ChatID)
		require.NoError(t, err)
		assert.Equal(t, ah.Username, saved.Username)
		assert.Equal(t, ah.Hash, saved.Hash)
		assert.Equal(t, ah.PeerID, saved.PeerID)
	})
}

func TestGetAccessHash(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	t.Run("get existing access hash", func(t *testing.T) {
		ah := entity.AccessHash{
			ChatID:   "-1003333333333",
			Username: "testbot3",
			Hash:     333333333,
			PeerID:   3333333333,
		}

		t.Cleanup(func() {
			err := elysiumRepo.DeleteAccessHash(context.Background(), ah.ChatID)
			require.NoError(t, err)
		})

		err := elysiumRepo.CreateOrUpdateAccessHash(context.Background(), ah)
		require.NoError(t, err)

		saved, err := elysiumRepo.GetAccessHash(context.Background(), ah.ChatID)
		require.NoError(t, err)
		assert.Equal(t, ah.ChatID, saved.ChatID)
		assert.Equal(t, ah.Username, saved.Username)
		assert.Equal(t, ah.Hash, saved.Hash)
		assert.Equal(t, ah.PeerID, saved.PeerID)
	})

	t.Run("get non-existing access hash", func(t *testing.T) {
		_, err := elysiumRepo.GetAccessHash(context.Background(), "-1009999999999")
		require.Error(t, err)
	})
}

func TestGetAllAccessHashes(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	t.Run("get all access hashes", func(t *testing.T) {
		hashes := []entity.AccessHash{
			{
				ChatID:   "-1004444444444",
				Username: "testbot4",
				Hash:     444444444,
				PeerID:   4444444444,
			},
			{
				ChatID:   "-1005555555555",
				Username: "testbot5",
				Hash:     555555555,
				PeerID:   5555555555,
			},
		}

		for _, ah := range hashes {
			t.Cleanup(func() {
				err := elysiumRepo.DeleteAccessHash(context.Background(), ah.ChatID)
				require.NoError(t, err)
			})

			err := elysiumRepo.CreateOrUpdateAccessHash(context.Background(), ah)
			require.NoError(t, err)
		}

		saved, err := elysiumRepo.GetAllAccessHashes(context.Background())
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(saved), len(hashes))

		foundCount := 0
		for _, ah := range hashes {
			for _, s := range saved {
				if s.ChatID == ah.ChatID {
					assert.Equal(t, ah.Username, s.Username)
					assert.Equal(t, ah.Hash, s.Hash)
					assert.Equal(t, ah.PeerID, s.PeerID)
					foundCount++
				}
			}
		}
		assert.Equal(t, len(hashes), foundCount)
	})

	t.Run("get all access hashes with empty table", func(t *testing.T) {
		hashes, err := elysiumRepo.GetAllAccessHashes(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, hashes)
	})
}

func TestDeleteAccessHash(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	t.Run("delete existing access hash", func(t *testing.T) {
		ah := entity.AccessHash{
			ChatID:   "-1006666666666",
			Username: "testbot6",
			Hash:     666666666,
			PeerID:   6666666666,
		}

		err := elysiumRepo.CreateOrUpdateAccessHash(context.Background(), ah)
		require.NoError(t, err)

		err = elysiumRepo.DeleteAccessHash(context.Background(), ah.ChatID)
		require.NoError(t, err)

		// Проверяем что запись удалена
		_, err = elysiumRepo.GetAccessHash(context.Background(), ah.ChatID)
		require.Error(t, err)
	})

	t.Run("delete non-existing access hash", func(t *testing.T) {
		err := elysiumRepo.DeleteAccessHash(context.Background(), "-1007777777777")
		require.Error(t, err)
	})
}
