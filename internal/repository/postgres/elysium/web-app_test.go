package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSaveWebAppEventSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	event := entity.WebAppEvent{
		EventType: entity.EventTypeInitApp,
		//UserID:         1,
		TelegramID: 6871255048,
		Payload:    json.RawMessage(`{"key":"value"}`),
		SessionID:  "session_123",
		Timestamp:  time.Now(),
	}

	err := elysiumRepo.SaveWebAppEvent(context.Background(), event)
	require.NoError(t, err)
}

func TestSaveWebAppEventFailsOnInvalidPayload(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	event := entity.WebAppEvent{
		EventType:  "test_event",
		TelegramID: 123456789,
		Payload:    json.RawMessage(`{"key": value}`), // Invalid JSON
		SessionID:  "session_123",
		Timestamp:  time.Now(),
	}

	err := elysiumRepo.SaveWebAppEvent(context.Background(), event)
	require.Error(t, err)
}

func TestSaveWebAppEventFailsOnEmptyEventType(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	event := entity.WebAppEvent{
		EventType:  "",
		TelegramID: 123456789,
		Payload:    json.RawMessage(`{"key":"value"}`),
		SessionID:  "session_123",
		Timestamp:  time.Now(),
	}

	err := elysiumRepo.SaveWebAppEvent(context.Background(), event)
	require.Error(t, err)
}

func TestSaveMultipleWebAppEventsSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	events := []entity.WebAppEvent{
		{
			EventType:  entity.EventTypeCloseAction,
			TelegramID: 123456789,
			Payload:    json.RawMessage(`{"key":"value1"}`),
			SessionID:  "session_123",
			Timestamp:  time.Now(),
		},
		{
			EventType:  entity.EventTypeExpandApp,
			TelegramID: 987654321,
			Payload:    json.RawMessage(`{"key":"value2"}`),
			SessionID:  "session_456",
			Timestamp:  time.Now(),
		},
	}

	err := elysiumRepo.SaveWebAppEvents(context.Background(), events)
	require.NoError(t, err)
}

func TestSaveWebAppEventsFailsOnEmptyEventList(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	err := elysiumRepo.SaveWebAppEvents(context.Background(), []entity.WebAppEvent{})
	require.Error(t, err)
	t.Log(err)
}

func TestSaveWebAppEventsFailsOnInvalidPayload(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	events := []entity.WebAppEvent{
		{
			EventType:  "event_type",
			TelegramID: 123456789,
			Payload:    json.RawMessage(`{"key": value}`), // Invalid JSON
			SessionID:  "session_123",
			Timestamp:  time.Now(),
		},
	}

	err := elysiumRepo.SaveWebAppEvents(context.Background(), events)
	require.Error(t, err)
	t.Log(err)
}
