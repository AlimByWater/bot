package demethra

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type UserEvent struct {
	UserID    int64
	EventType string
	Timestamp time.Time
	Data      map[string]interface{}
}

type Repository interface {
	SaveUserEvent(ctx context.Context, event UserEvent) error
	GetUserEvents(ctx context.Context, userID int64, since time.Time) ([]UserEvent, error)
}

type Module struct {
	repo   Repository
	logger *slog.Logger
}

func NewModule(repo Repository, logger *slog.Logger) *Module {
	return &Module{
		repo:   repo,
		logger: logger,
	}
}

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
	case entity.EventTypeResumeAnimation:
		return m.handleResumeAnimation(ctx, event)
	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}
}

func (m *Module) saveUserEvent(ctx context.Context, userID int64, eventType string, data map[string]interface{}) error {
	event := UserEvent{
		UserID:    userID,
		EventType: eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	err := m.repo.SaveUserEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to save user event to repository: %w", err)
	}
	return nil
}

func (m *Module) getUserState(ctx context.Context, userID int64) (map[string]interface{}, error) {
	events, err := m.repo.GetUserEvents(ctx, userID, time.Now().Add(-24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("failed to get user events from repository: %w", err)
	}

	state := make(map[string]interface{})
	for _, event := range events {
		for k, v := range event.Data {
			state[k] = v
		}
	}
	return state, nil
}

func (m *Module) handleInitialization(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Web app initialized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveUserEvent(ctx, event.TelegramUserID, "initialization", nil)
}

func (m *Module) handleClosing(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Web app closed", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveUserEvent(ctx, event.TelegramUserID, "closing", nil)
}

func (m *Module) handleStartRadio(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Radio started", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveUserEvent(ctx, event.TelegramUserID, "start_radio", map[string]interface{}{"isRadioPlaying": true})
}

func (m *Module) handleStartAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation started", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveUserEvent(ctx, event.TelegramUserID, "start_animation", map[string]interface{}{"isAnimationActive": true})
}

func (m *Module) handleMinimize(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("App minimized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	if err := m.saveUserEvent(ctx, event.TelegramUserID, "minimize", map[string]interface{}{"isMinimized": true}); err != nil {
		return err
	}
	return m.handlePauseAnimation(ctx, event)
}

func (m *Module) handleMaximize(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("App maximized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	if err := m.saveUserEvent(ctx, event.TelegramUserID, "maximize", map[string]interface{}{"isMinimized": false}); err != nil {
		return err
	}
	return m.handleResumeAnimation(ctx, event)
}

func (m *Module) handlePauseAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation paused", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveUserEvent(ctx, event.TelegramUserID, "pause_animation", map[string]interface{}{"isAnimationActive": false})
}

func (m *Module) handleResumeAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation resumed", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	return m.saveUserEvent(ctx, event.TelegramUserID, "resume_animation", map[string]interface{}{"isAnimationActive": true})
}
