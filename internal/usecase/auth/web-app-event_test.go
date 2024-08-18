package auth

import (
	"arimadj-helper/internal/entity"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitAppPayload(t *testing.T) {
	rawJson := []byte(`{
    "event_type": "init_app",
    "session_id": "0190f960-5bc4-7008-97c2-6ed003045d0b",
    "telegram_user_id": 464230212,
    "payload": {
    },
    "timestamp": "0001-01-01T00:00:00Z"
}`)

	var event entity.WebAppEvent
	err := json.Unmarshal(rawJson, &event)
	require.NoError(t, err)

	//payloadMap, ok := event.Payload
	//require.Equal(t, true, ok)

	if event.Payload != nil {
		var payload entity.InitAppPayload
		err = json.Unmarshal(event.Payload, &payload)
		require.NoError(t, err)
		t.Log(payload.RawInitData)
	} else {
		t.Log("payload empty")
	}

	//err = initdata.Validate(payload.RawInitData, "7486051673:AAGXMsNZ3ia99ljU48IErrA5PH4ZV-VncFo", 24*time.Hour)
	//require.NoError(t, err)
	//
	//parsedData, err := initdata.Parse(payload.RawInitData)
	//require.NoError(t, err)
	//
	//j, _ := json.MarshalIndent(parsedData, "", "  ")
	//t.Log(string(j))

}
