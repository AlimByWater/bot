package demethra

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"log/slog"
	"time"
)

func (m *Module) ProcessWebAppEvent(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info(fmt.Sprintf("Received WebAppEvent: Type=%s, UserID=%d, SessionID=%s", event.EventType, event.TelegramUserID, event.SessionID))

	switch event.EventType {
	case entity.EventTypeInitApp:
		return m.handleInitialization(ctx, event)
	case entity.EventTypeCloseApp:
		return m.handleClosing(ctx, event)
	case entity.EventTypeStartRadio:
		return m.handleStartRadio(ctx, event)
	case entity.EventTypeStartAnimation:
		return m.handleStartAnimation(ctx, event)
	case entity.EventTypeMinimizeApp:
		return m.handleMinimize(ctx, event)
	case entity.EventTypeMaximizeApp:
		return m.handleMaximize(ctx, event)
	case entity.EventTypePauseAnimation:
		return m.handlePauseAnimation(ctx, event)
	case entity.EventTypeCloseAnimation:
		return m.handleCloseAnimation(ctx, event)
	case entity.EventTypeResumeAnimation:
		return m.handleResumeAnimation(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

func (m *Module) saveWebAppEvent(ctx context.Context, event entity.WebAppEvent) error {
	event.Timestamp = time.Now()
	
	var payloadID int64
	var additionalData map[string]interface{}

	switch event.EventType {
	case entity.EventTypeStartRadio:
		payload := event.Payload.(map[string]interface{})
		radioEvent := RadioEvent{
			Duration: int(payload["duration"].(float64)),
			TrackID:  payload["trackID"].(string),
		}
		var err error
		payloadID, err = m.repo.SaveRadioEvent(ctx, radioEvent)
		if err != nil {
			return fmt.Errorf("failed to save radio event: %w", err)
		}
	case entity.EventTypeStartAnimation, entity.EventTypePauseAnimation, entity.EventTypeResumeAnimation, entity.EventTypeCloseAnimation:
		payload := event.Payload.(map[string]interface{})
		animationEvent := AnimationEvent{
			AnimationID: payload["animationID"].(string),
			Duration:    int(payload["duration"].(float64)),
		}
		var err error
		payloadID, err = m.repo.SaveAnimationEvent(ctx, animationEvent)
		if err != nil {
			return fmt.Errorf("failed to save animation event: %w", err)
		}
	default:
		additionalData = event.Payload.(map[string]interface{})
	}

	err := m.repo.SaveWebAppEvent(ctx, WebAppEvent{
		EventType:      event.EventType,
		UserID:         event.UserID,
		TelegramUserID: event.TelegramUserID,
		SessionID:      event.SessionID,
		Timestamp:      event.Timestamp,
		PayloadID:      payloadID,
		AdditionalData: additionalData,
	})
	if err != nil {
		return fmt.Errorf("failed to save web app event: %w", err)
	}
	return nil
}

func (m *Module) handleInitialization(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Web app initialized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleClosing(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Web app closed", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleStartRadio(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Radio started", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleStartAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation started", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleMinimize(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("App minimized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleMaximize(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("App maximized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handlePauseAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation paused", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleResumeAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation resumed", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveWebAppEvent(ctx, event)
}

func (m *Module) handleCloseAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation closed", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveWebAppEvent(ctx, event)
}
