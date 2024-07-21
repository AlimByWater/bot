package elysium_test

import (
	"arimadj-helper/internal/entity"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSaveWebAppEventSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	event := entity.WebAppEvent{
		EventType: entity.EventTypeResumeAnimation,
		//UserID:         1,
		TelegramUserID: 6871255048,
		Payload:        json.RawMessage(`{"key":"value"}`),
		SessionID:      "session_123",
		Timestamp:      time.Now(),
	}

	err := elysiumRepo.SaveWebAppEvent(context.Background(), event)
	require.NoError(t, err)
}

func TestSaveWebAppEventFailsOnInvalidPayload(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	event := entity.WebAppEvent{
		EventType:      "test_event",
		UserID:         1,
		TelegramUserID: 123456789,
		Payload:        json.RawMessage(`{"key": value}`), // Invalid JSON
		SessionID:      "session_123",
		Timestamp:      time.Now(),
	}

	err := elysiumRepo.SaveWebAppEvent(context.Background(), event)
	require.Error(t, err)
}

func TestSaveWebAppEventFailsOnEmptyEventType(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	event := entity.WebAppEvent{
		EventType:      "",
		UserID:         1,
		TelegramUserID: 123456789,
		Payload:        json.RawMessage(`{"key":"value"}`),
		SessionID:      "session_123",
		Timestamp:      time.Now(),
	}

	err := elysiumRepo.SaveWebAppEvent(context.Background(), event)
	require.Error(t, err)
}
