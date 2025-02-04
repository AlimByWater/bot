package userbot

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/dispatcher/handlers/filters"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/functions"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/glebarez/sqlite"
	"github.com/go-telegram/bot"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
	"log/slog"
	"math"
	"strconv"
	"strings"
	"sync"
)

type repo interface {
	CreateOrUpdateAccessHash(ctx context.Context, accessHash entity.AccessHash) error
	GetAccessHash(ctx context.Context, chatID string) (entity.AccessHash, error)
	GetAllAccessHashes(ctx context.Context) ([]entity.AccessHash, error)
}

type config interface {
	GetUserBotAppID() int
	GetUserBotAppHash() string
	GetUserBotTgPhone() string
	GetSessionDir() string
}

type User struct {
	client     *gotgproto.Client
	ctx        context.Context
	logger     *slog.Logger
	cfg        config
	accessHash sync.Map
	repo       repo
	id         int64
}

func NewBot(repo repo, cfg config) *User {
	u := &User{
		repo:       repo,
		cfg:        cfg,
		accessHash: sync.Map{},
	}
	return u
}

func (u *User) Init(ctx context.Context, logger *slog.Logger) error {
	u.ctx = ctx
	u.logger = logger.With(slog.String("module", "USERBOT"))

	appID := u.cfg.GetUserBotAppID()

	appHash := u.cfg.GetUserBotAppHash()
	if appHash == "" {
		return fmt.Errorf("не указан APP_HASH")
	}

	phone := u.cfg.GetUserBotTgPhone()
	if phone == "" {
		return fmt.Errorf("не указан номер телефона TG_PHONE")
	}

	//var err2 error
	client, err := gotgproto.NewClient(
		appID,
		appHash,
		gotgproto.ClientTypePhone(phone),
		&gotgproto.ClientOpts{
			Session: sessionMaker.SqlSession(sqlite.Open(u.cfg.GetSessionDir() + "/session.db")),
		},
	)
	if err != nil {
		return fmt.Errorf("ошибка создания клиента: %v", err)
	}

	u.client = client

	// regex for message starting with /emoji
	//emojiCmd, err := filters.Message.Regex("^/emoji")
	//if err != nil {
	//	return fmt.Errorf("ошибка создания regex: %v", err)
	//}

	//dispatcher.AddHandlerToGroup(handlers.NewMessage(emojiCmd, u.emoji), 0)

	//dispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Message.Text, u.echo), 0)
	u.client.Dispatcher.AddHandlerToGroup(handlers.NewMessage(filters.Message.Text, u.saveAccessHash), 0) // сохраняем хэш доступа в базу данных

	self, err := u.client.Client.Self(ctx)
	if err != nil {
		return fmt.Errorf("ошибка получения себя: %v", err)
	}

	u.id = self.ID

	accessHashes, err := u.repo.GetAllAccessHashes(ctx)
	if err != nil {
		return fmt.Errorf("getAllAccessHashes: %v", err)
	}

	for _, accessHash := range accessHashes {
		u.accessHash.Store(accessHash.PeerID, accessHash)
	}

	return nil
}

func (u *User) GetID() int64 {
	return u.id
}

func (u *User) DeleteMessage(ctx context.Context, msgID int) error {
	sender := message.NewSender(tg.NewClient(u.client))

	//id, ok := u.chatIdsToInternalIds.Load(chatID)
	//if !ok {
	//	return fmt.Errorf("id не найден доступ к чату %s", chatID)
	//}
	//
	//ah, ok := u.accessHash.Load(id)
	//if !ok {
	//	return fmt.Errorf("ah не найден доступ к чату %s", chatID)
	//}
	//peer := &tg.InputPeerChannel{
	//	ChannelID:  id.(int64),
	//	AccessHash: ah.(int64),
	//}

	_, err := sender.Delete().Messages(ctx, msgID)
	//.Delete().Messages(ctx, msgID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

func (u *User) SendMessage(ctx context.Context, chatID string, msg bot.SendMessageParams) error {
	sender := message.NewSender(tg.NewClient(u.client))

	ah, err := u.chatIDToAccessHash(chatID)
	if err != nil {
		return fmt.Errorf("ошибка получения доступа к чату: %v", err)
	}

	peer := &tg.InputPeerChannel{
		ChannelID:  ah.PeerID,
		AccessHash: ah.Hash,
	}

	var formats []message.StyledTextOption
	formats = append(formats,
		styling.Plain(msg.Text),
		styling.Plain("\n"))

	_, err = sender.To(peer).SendAs(peer).Reply(msg.ReplyParameters.MessageID).StyledText(ctx, formats...)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %v", err)
	}

	return nil
}

func (u *User) SendMessageWithEmojis(ctx context.Context, chatID string, width int, packLink string, emojis []entity.EmojiMeta, replyTo int) error {
	sender := message.NewSender(tg.NewClient(u.client))

	ah, err := u.chatIDToAccessHash(chatID)
	if err != nil {
		return fmt.Errorf("ошибка получения доступа к чату: %v", err)
	}
	peer := &tg.InputPeerChannel{
		ChannelID:  ah.PeerID,
		AccessHash: ah.Hash,
	}

	formats, err := u.styledText(width, packLink, emojis)
	if err != nil {
		return fmt.Errorf("ошибка форматирования текста: %v", err)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	_, err = sender.To(peer).SendAs(peer).Reply(replyTo).NoWebpage().StyledText(ctx, formats...)
	if err != nil {
		return fmt.Errorf("ошибка отправки сообщения: %v", err)
	}

	return nil
}

func (u *User) styledText(width int, packLink string, emojis []entity.EmojiMeta) ([]message.StyledTextOption, error) {
	var formats []message.StyledTextOption

	if width < entity.DefaultWidth {
		width = entity.DefaultWidth
	}

	// "⠀"
	for i, emoji := range emojis {
		if i == len(emojis)-1 || i == entity.MaxStickerInMessage-1 {
			break
		}
		if emoji.Transparent {
			formats = append(formats, styling.Plain("⠀⠀"))
		} else {
			documentID, err := strconv.ParseInt(emoji.DocumentID, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("ошибка при парсинге id документа: %v", err)
			}
			formats = append(formats, styling.CustomEmoji("⭐️", documentID))
		}
		if math.Mod(float64(i+1), float64(width)) == 0 {
			formats = append(formats, styling.Plain("\n"))
		}
	}

	formats = append(formats,
		styling.Plain("\t"),
		styling.TextURL("⁂добавить", fmt.Sprintf("https://t.me/addemoji/%s", packLink)),
	)

	return formats, nil
}

func (u *User) SendMessageWithEmojisToBot(ctx context.Context, chatID string, width int, packLink string, emojis []entity.EmojiMeta) (int, error) {
	sender := message.NewSender(tg.NewClient(u.client))

	ah, err := u.chatIDToAccessHash(chatID)
	if err != nil {
		return 0, fmt.Errorf("ошибка получения доступа к чату: %v", err)
	}
	peer := &tg.InputPeerUser{
		UserID:     ah.PeerID,
		AccessHash: ah.Hash,
	}

	formats, err := u.styledText(width, packLink, emojis)
	if err != nil {
		return 0, fmt.Errorf("ошибка форматирования текста: %v", err)
	}

	updClass, err := sender.To(peer).NoWebpage().StyledText(ctx, formats...)
	if err != nil {
		return 0, fmt.Errorf("ошибка отправки сообщения: %v", err)
	}

	var tgMsg *tg.Message
	tgMsg = functions.GetNewMessageUpdate(tgMsg, updClass, u.client.PeerStorage)

	upd, ok := updClass.(*tg.Updates)
	if !ok {
		return 0, fmt.Errorf("ошибка преобразования объекта: %v", err)
	}

	updMessage, ok := upd.Updates[0].(*tg.UpdateMessageID)
	if !ok {
		return 0, fmt.Errorf("ошибка преобразования объекта: %v", err)
	}

	return updMessage.ID, nil
}

func (u *User) Shutdown(ctx context.Context) {
	u.logger.Info("Завершение работы User...")
	u.client.Stop()
	u.logger.Info("User остановлен")
}

func (u *User) echo(ctx *ext.Context, update *ext.Update) error {
	//u.logger.Info("input peer", update.EffectiveChat().GetInputChannel(), update.EffectiveChat().GetInputPeer(), update.UpdateClass.TypeName())
	//u.logger.Info("message", update.EffectiveMessage.ID, update.EffectiveMessage.Text)
	//u.logger.Info("user", update.GetUserChat().Username)

	select {
	case <-ctx.Done():
		return nil
	default:
		sender := message.NewSender(tg.NewClient(u.client))
		peer := u.client.PeerStorage.GetInputPeerById(update.EffectiveChat().GetID())

		u.logger.Info("input peer", slog.String("peer", peer.String()))

		_, err := sender.To(peer).Reply(update.EffectiveMessage.ID).Text(ctx, update.EffectiveMessage.Text)
		if err != nil {
			u.logger.Error("Failed to send message by userBot", slog.String("err", err.Error()))
		}
	}
	return nil
}

// saveAccessHash сохраняет все хэши доступа от чатов с ботами в базу данных
func (u *User) saveAccessHash(ctx *ext.Context, update *ext.Update) error {
	userChat := update.GetUserChat()
	if userChat == nil {
		//u.logger.Error("update.GetUserChat() вернул nil")
		return nil
	}
	if !userChat.Bot {
		return nil
	}

	effectiveChat := update.EffectiveChat()
	if effectiveChat == nil {
		u.logger.Error("update.EffectiveChat() вернул nil")
		return nil
	}

	chatID := fmt.Sprintf("-100%d", effectiveChat.GetID())

	accessHash := entity.AccessHash{
		Username: userChat.Username,
		ChatID:   chatID,
		Hash:     effectiveChat.GetAccessHash(),
		PeerID:   effectiveChat.GetID(),
	}

	u.accessHash.Store(accessHash.PeerID, accessHash)

	err := u.repo.CreateOrUpdateAccessHash(ctx, accessHash)
	if err != nil {
		u.logger.Error("Failed to save access hash", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (u *User) chatIDToAccessHash(chatID string) (entity.AccessHash, error) {
	chatID = strings.TrimPrefix(chatID, "-100")
	peerID, err := strconv.ParseInt(chatID, 10, 64)
	if err != nil {
		return entity.AccessHash{}, fmt.Errorf("ошибка парсинга id чата: %v", err)
	}

	ahRaw, ok := u.accessHash.Load(peerID)
	if !ok {
		return entity.AccessHash{}, fmt.Errorf("не найден доступ к чату %s", chatID)
	}

	ah, ok := ahRaw.(entity.AccessHash)
	if !ok {
		return entity.AccessHash{}, fmt.Errorf("ошибка преобразования типов %s", chatID)
	}

	return ah, nil
}
