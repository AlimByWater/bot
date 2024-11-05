package demethra

import (
	"context"
	"elysium/internal/entity"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

func (m *Module) ProcessWebAppEvent(ctx context.Context, event entity.WebAppEvent) {
	//m.logger.Debug(fmt.Sprintf("Received WebAppEvent: Type=%s, TelegramUserID=%d, SessionID=%s", event.EventType, event.TelegramID, event.SessionID))
	var err error
	switch event.EventType {
	case entity.EventTypeInitApp:
		err = m.handleInitialization(ctx, event)
	case entity.EventTypeStartApp:
		err = m.handleStartApp(ctx, event)
	case entity.EventTypeStartAction:
		err = m.handleStartAction(ctx, event)
	case entity.EventTypeCollapseApp:
		err = m.handleMinimize(ctx, event)
	case entity.EventTypeExpandApp:
		err = m.handleMaximize(ctx, event)
	case entity.EventTypeCloseAction:
		err = m.handleCloseAction(ctx, event)
	case entity.EventTypeChangedStream:
		err = m.handleChangedStream(ctx, event)
	default:
		err = fmt.Errorf("unknown event type: %s", event.EventType)
	}

	if err != nil {
		m.logger.Error("Failed to process web app event", slog.String("error", err.Error()), slog.Any("event", event), slog.String("method", "ProcessWebAppEvent"))
	}
}

func (m *Module) saveWebAppEvent(ctx context.Context, event entity.WebAppEvent) error {
	event.Timestamp = time.Now()

	m.batchEventUpdate <- event
	return nil
}

func (m *Module) batchUpdateLoop() {
	t := time.NewTicker(10 * time.Second)
	for {
		var events []entity.WebAppEvent
		select {
		case <-m.ctx.Done():
			return
		case event := <-m.batchEventUpdate:
			events = append(events, event)
			if len(events) >= batchItemsCount {
				err := m.repo.SaveWebAppEvents(m.ctx, events)
				if err != nil {
					m.logger.Error(fmt.Sprintf("failed to save web app events: %v", err), slog.String("method", "batchUpdateLoop"))
				}
				events = nil
			}
		case <-t.C: // flush events every 10 seconds
			if len(events) > 0 {
				err := m.repo.SaveWebAppEvents(m.ctx, events)
				if err != nil {
					m.logger.Error(fmt.Sprintf("failed to save web app events: %v", err), slog.String("method", "batchUpdateLoop"))
				}
				events = nil
			}
		}

		for i := 0; i < batchItemsCount; i++ {
			var event entity.WebAppEvent
			select {
			case event = <-m.batchEventUpdate:
				events = append(events, event)
				continue
			default:
			}

			if len(events) == 0 {
				select {
				case <-m.ctx.Done():
					// return if we processed all messages from channel and it's the graceful shutdown case
					return
				default:
					time.Sleep(time.Second) // do not waste cpu on idle loop iterations
					continue
				}
			}
		}
	}
}

func (m *Module) handleInitialization(ctx context.Context, event entity.WebAppEvent) error {
	var payload entity.InitAppPayload
	if event.Payload != nil {
		err := json.Unmarshal(event.Payload, &payload)
		if err != nil {
			return fmt.Errorf("invalid payload for init event")
		}
	}

	err := initdata.Validate(payload.RawInitData, m.cfg.GetBotToken(), 24*time.Hour)
	if err != nil {
		err2 := initdata.Validate(payload.RawInitData, "7486051673:AAEg2bzMqec1NkFK8tHycLn8gvGxK6xQ6ww", 24*time.Hour)
		if err2 != nil {
			return fmt.Errorf("invalid init data: %w: %w", err, err2)
		}
	}

	parsedData, err := initdata.Parse(payload.RawInitData)
	if err != nil {
		return fmt.Errorf("failed to parse init data: %w", err)
	}

	_, err = m.cache.GetListenerByTelegramID(ctx, parsedData.User.ID)
	if err != nil {
		listenerCache := entity.ListenerCache{
			TelegramID: parsedData.User.ID,
			Payload: entity.ListenerCachePayload{
				InitTimestamp: time.Now().Unix(),
				LastActivity:  time.Now().Unix(),
			},
		}

		err = m.cache.SaveOrUpdateListener(ctx, listenerCache)
		if err != nil {
			m.logger.Error(fmt.Sprintf("failed to save listener: %v", err), slog.String("method", "handleInitialization"))
		}
	}

	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleChangedStream(ctx context.Context, event entity.WebAppEvent) error {
	var payload entity.ChangedStreamPayload
	if event.Payload != nil {
		err := json.Unmarshal(event.Payload, &payload)
		if err != nil {
			return fmt.Errorf("invalid payload for changed stream event: %w", err)
		}
	}
	err := m.cache.SaveOrUpdateListener(ctx, entity.ListenerCache{
		TelegramID: event.TelegramID,
		Payload: entity.ListenerCachePayload{
			StreamSlug: payload.StreamSlug,
		},
	})
	if err != nil {
		m.logger.Error("Failed to save or update listener in cache", slog.String("error", err.Error()), slog.Int64("telegram_id", event.TelegramID), slog.String("method", "handleChangedStream"))
	}

	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleStartApp(ctx context.Context, event entity.WebAppEvent) error {
	return m.saveWebAppEvent(ctx, event)
}
func (m *Module) handleStartAction(ctx context.Context, event entity.WebAppEvent) error {
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleMinimize(ctx context.Context, event entity.WebAppEvent) error {
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleMaximize(ctx context.Context, event entity.WebAppEvent) error {
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleCloseAction(ctx context.Context, event entity.WebAppEvent) error {
	return m.saveWebAppEvent(ctx, event)
}
