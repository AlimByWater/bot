package demethra

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type userState struct {
	isRadioPlaying    bool
	isAnimationActive bool
	isMinimized       bool
	lastUpdated       time.Time
}

type Repository interface {
	SaveUserState(ctx context.Context, userID int64, state *userState) error
	GetUserState(ctx context.Context, userID int64) (*userState, error)
}

type Module struct {
	// ... existing fields ...
	userStates map[int64]*userState
	stateMutex sync.RWMutex
	repo       Repository
	logger     *slog.Logger
}

func NewModule(repo Repository, logger *slog.Logger) *Module {
	m := &Module{
		userStates: make(map[int64]*userState),
		repo:       repo,
		logger:     logger,
	}
	go m.startCacheCleaner()
	return m
}

func (m *Module) startCacheCleaner() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		m.cleanCache()
	}
}

func (m *Module) cleanCache() {
	m.stateMutex.Lock()
	defer m.stateMutex.Unlock()

	for userID, state := range m.userStates {
		if time.Since(state.lastUpdated) > 1*time.Hour {
			delete(m.userStates, userID)
		}
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

func (m *Module) getUserState(ctx context.Context, userID int64) (*userState, error) {
	m.stateMutex.RLock()
	state, exists := m.userStates[userID]
	m.stateMutex.RUnlock()

	if !exists {
		var err error
		state, err = m.repo.GetUserState(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user state from repository: %w", err)
		}
		if state == nil {
			state = &userState{}
		}
		m.stateMutex.Lock()
		m.userStates[userID] = state
		m.stateMutex.Unlock()
	}

	return state, nil
}

func (m *Module) saveUserState(ctx context.Context, userID int64, state *userState) error {
	state.lastUpdated = time.Now()
	m.stateMutex.Lock()
	m.userStates[userID] = state
	m.stateMutex.Unlock()

	err := m.repo.SaveUserState(ctx, userID, state)
	if err != nil {
		return fmt.Errorf("failed to save user state to repository: %w", err)
	}
	return nil
}

func (m *Module) handleInitialization(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Web app initialized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state := &userState{} // Reset state
	return m.saveUserState(ctx, event.TelegramUserID, state)
}

func (m *Module) handleClosing(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Web app closed", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	m.stateMutex.Lock()
	delete(m.userStates, event.TelegramUserID)
	m.stateMutex.Unlock()
	return m.repo.SaveUserState(ctx, event.TelegramUserID, nil) // Clear state in repository
}

func (m *Module) handleStartRadio(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Radio started", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state, err := m.getUserState(ctx, event.TelegramUserID)
	if err != nil {
		return err
	}
	state.isRadioPlaying = true
	// TODO: Implement actual radio start logic for this user
	return m.saveUserState(ctx, event.TelegramUserID, state)
}

func (m *Module) handleStartAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation started", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state, err := m.getUserState(ctx, event.TelegramUserID)
	if err != nil {
		return err
	}
	state.isAnimationActive = true
	// TODO: Implement actual animation start logic for this user
	return m.saveUserState(ctx, event.TelegramUserID, state)
}

func (m *Module) handleMinimize(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("App minimized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state, err := m.getUserState(ctx, event.TelegramUserID)
	if err != nil {
		return err
	}
	state.isMinimized = true
	if err := m.saveUserState(ctx, event.TelegramUserID, state); err != nil {
		return err
	}
	return m.handlePauseAnimation(ctx, event)
}

func (m *Module) handleMaximize(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("App maximized", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state, err := m.getUserState(ctx, event.TelegramUserID)
	if err != nil {
		return err
	}
	state.isMinimized = false
	if err := m.saveUserState(ctx, event.TelegramUserID, state); err != nil {
		return err
	}
	return m.handleResumeAnimation(ctx, event)
}

func (m *Module) handlePauseAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation paused", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state, err := m.getUserState(ctx, event.TelegramUserID)
	if err != nil {
		return err
	}
	state.isAnimationActive = false
	// TODO: Implement actual animation pause logic for this user
	return m.saveUserState(ctx, event.TelegramUserID, state)
}

func (m *Module) handleResumeAnimation(ctx context.Context, event entity.WebAppEvent) error {
	m.logger.Info("Animation resumed", slog.String("sessionID", event.SessionID), slog.Int64("userID", event.TelegramUserID))
	state, err := m.getUserState(ctx, event.TelegramUserID)
	if err != nil {
		return err
	}
	state.isAnimationActive = true
	// TODO: Implement actual animation resume logic for this user
	return m.saveUserState(ctx, event.TelegramUserID, state)
}
