package demethra

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type userState struct {
	isRadioPlaying   bool
	isAnimationActive bool
	isMinimized       bool
}

type Module struct {
	// ... existing fields ...
	userStates map[int64]*userState
	stateMutex sync.RWMutex
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

func (m *Module) getUserState(userID int64) *userState {
	m.stateMutex.RLock()
	state, exists := m.userStates[userID]
	m.stateMutex.RUnlock()

	if !exists {
		m.stateMutex.Lock()
		state = &userState{}
		m.userStates[userID] = state
		m.stateMutex.Unlock()
	}

	return state
}

func (m *Module) handleInitialization(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Web app initialized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state := m.getUserState(event.TelegramUserID)
	*state = userState{} // Reset state
	return nil
}

func (m *Module) handleClosing(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Web app closed", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	m.stateMutex.Lock()
	delete(m.userStates, event.TelegramUserID)
	m.stateMutex.Unlock()
	return nil
}

func (m *Module) handleStartRadio(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Radio started", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state := m.getUserState(event.TelegramUserID)
	state.isRadioPlaying = true
	// TODO: Implement actual radio start logic for this user
	return nil
}

func (m *Module) handleStartAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation started", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state := m.getUserState(event.TelegramUserID)
	state.isAnimationActive = true
	// TODO: Implement actual animation start logic for this user
	return nil
}

func (m *Module) handleMinimize(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("App minimized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state := m.getUserState(event.TelegramUserID)
	state.isMinimized = true
	return m.handlePauseAnimation(ctx, event)
}

func (m *Module) handleMaximize(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("App maximized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state := m.getUserState(event.TelegramUserID)
	state.isMinimized = false
	return m.handleResumeAnimation(ctx, event)
}

func (m *Module) handlePauseAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation paused", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state := m.getUserState(event.TelegramUserID)
	state.isAnimationActive = false
	// TODO: Implement actual animation pause logic for this user
	return nil
}

func (m *Module) handleResumeAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation resumed", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state := m.getUserState(event.TelegramUserID)
	state.isAnimationActive = true
	// TODO: Implement actual animation resume logic for this user
	return nil
}
