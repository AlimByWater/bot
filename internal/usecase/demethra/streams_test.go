package demethra_test

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGetStreamsMetaInfo(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	t.Run("Success", func(t *testing.T) {
		time.Sleep(time.Second * 15)
		streams := module.GetStreamsMetaInfo()

		require.NotEmpty(t, streams)
		j, _ := streams.MarshalJSON()
		t.Log(string(j))
		t.Log(streams)
	})
}
