package demethra

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"log/slog"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
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
	err := m.repo.SaveWebAppEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("failed to save web app event: %w", err)
	}
	return nil
}

func (m *Module) handleInitialization(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Web app initialized", slog.String("sessionID", event.SessionID), slog.Int64("telegram_id", event.TelegramUserID))

	rawInitData, ok := event.Payload.(string)
	if !ok {
		return fmt.Errorf("invalid payload for init event")
	}

	err := initdata.Validate(rawInitData, m.cfg.GetBotToken(), 24*time.Hour)
	if err != nil {
		m.logger.Error("Invalid init data", slog.String("error", err.Error()))
		return fmt.Errorf("invalid init data: %w", err)
	}

	parsedData, err := initdata.Parse(rawInitData)
	if err != nil {
		m.logger.Error("Failed to parse init data", slog.String("error", err.Error()))
		return fmt.Errorf("failed to parse init data: %w", err)
	}

	// создаем или обновляем пользователя на всякий случай
	user, err := m.repo.CreateOrUpdateUser(ctx, entity.User{
		TelegramID:       parsedData.User.ID,
		TelegramUsername: parsedData.User.Username,
		Firstname:        parsedData.User.FirstName,
		DateCreate:       time.Now(),
	})
	if err != nil {
		m.logger.Debug(fmt.Sprintf("Failed to create or update user: %v", err), slog.String("method", "handleInitialization"))
	} else {
		m.logger.Debug(fmt.Sprintf("User created: id=%v; telegram_id=%v", user.ID, user.TelegramID), slog.String("method", "handleInitialization"))
	}

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
