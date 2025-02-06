package progress_manager

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"strconv"
	"sync"

	"github.com/mymmrac/telego"
)

// ProgressManager управляет сообщениями о прогрессе
type ProgressManager struct {
	progressMessages sync.Map
	cancelChannels   sync.Map // Добавляем мапу для каналов отмены
}

// NewManager создает новый менеджер прогресса
func NewManager() *ProgressManager {
	return &ProgressManager{
		progressMessages: sync.Map{},
		cancelChannels:   sync.Map{},
	}
}

// SendMessage отправляет новое сообщение о прогрессе
func (m *ProgressManager) SendMessage(ctx context.Context, b *telego.Bot, chatID telego.ChatID, replyToID int, userID int64, status string) (*entity.ProgressMessage, error) {
	// Формируем ключ отмены только из необходимых данных
	cancelKey := fmt.Sprintf("%d_%d_%d", chatID.ID, replyToID, userID)

	// Создаем кнопку отмены
	keyboard := &telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				{Text: "❌ Остановить генерацию", CallbackData: "cancel_" + cancelKey},
			},
		},
	}

	params := &telego.SendMessageParams{
		ReplyParameters: &telego.ReplyParameters{
			MessageID: replyToID,
			ChatID:    chatID,
		},
		ChatID:      chatID,
		Text:        status,
		ReplyMarkup: keyboard,
	}

	msg, err := b.SendMessage(params)
	if err != nil {
		return nil, fmt.Errorf("failed to send progress message: %w", err)
	}

	progress := &entity.ProgressMessage{
		ChatID:    chatID.ID,
		MessageID: msg.MessageID,
		Status:    status,
		CancelKey: cancelKey,
		UserID:    userID,
	}

	key := strconv.FormatInt(chatID.ID, 10) + ":" + strconv.Itoa(msg.MessageID)
	m.progressMessages.Store(key, progress)

	// Создаем канал отмены
	cancelCh := make(chan struct{})
	m.cancelChannels.Store(cancelKey, cancelCh)

	return progress, nil
}

// GetCancelChannel возвращает канал отмены
func (m *ProgressManager) GetCancelChannel(cancelKey string) chan struct{} {
	if ch, ok := m.cancelChannels.Load(cancelKey); ok {
		return ch.(chan struct{})
	}
	return nil
}

// Cancel отменяет процесс по ключу
func (m *ProgressManager) Cancel(cancelKey string) {
	if ch, ok := m.cancelChannels.Load(cancelKey); ok {
		close(ch.(chan struct{}))
		m.cancelChannels.Delete(cancelKey)
	}
}

// DeleteMessage удаляет сообщение о прогрессе
func (m *ProgressManager) DeleteMessage(ctx context.Context, b *telego.Bot, chatID telego.ChatID, msgID int) error {
	key := strconv.FormatInt(chatID.ID, 10) + ":" + strconv.Itoa(msgID)
	progressRaw, exists := m.progressMessages.Load(key)
	if !exists {
		return nil // Если сообщения нет, это не ошибка
	}

	progress := progressRaw.(*entity.ProgressMessage)
	params := &telego.DeleteMessageParams{
		ChatID:    chatID,
		MessageID: progress.MessageID,
	}

	err := b.DeleteMessage(params)
	if err != nil {
		return fmt.Errorf("failed to delete progress message: %w", err)
	}

	m.progressMessages.Delete(key)
	return nil
}

// UpdateMessage обновляет существующее сообщение о прогрессе
func (m *ProgressManager) UpdateMessage(ctx context.Context, b *telego.Bot, chatID telego.ChatID, msgID int, status string) error {
	key := strconv.FormatInt(chatID.ID, 10) + ":" + strconv.Itoa(msgID)
	progressRaw, exists := m.progressMessages.Load(key)
	if !exists {
		return fmt.Errorf("progress message not found for chat %d", chatID.ID)
	}

	progress := progressRaw.(*entity.ProgressMessage)

	// Сохраняем клавиатуру
	keyboard := &telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				{Text: "❌ Остановить генерацию", CallbackData: "cancel_" + progress.CancelKey},
			},
		},
	}

	params := &telego.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   progress.MessageID,
		Text:        status,
		ReplyMarkup: keyboard,
	}

	_, err := b.EditMessageText(params)
	if err != nil {
		return fmt.Errorf("failed to update progress message: %w", err)
	}

	progress.Status = status
	return nil
}
