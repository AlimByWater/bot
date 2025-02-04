package processing

import (
	"context"
	"log/slog"
	"sync"

	"elysium/internal/entity"
	"github.com/go-telegram/bot/models"
)

type ModuleInterface interface {
	RoundDimensions(width, height int) (int, int)
	RoundUpTo100(num int) int
	DimensionToNewWidth(width, height, newWidth int) (int, int)
	getVideoDimensions(inputVideo string) (width, height int, err error)
	ProcessVideo(ctx context.Context, args *entity.EmojiCommand) ([]string, error)
	RegisterDirectory(dir string) error
	CheckAndRemoveOldDirectories()
	ExtractCommandArgs(msgText, msgCaption string) string
	SetupEmojiCommand(args *entity.EmojiCommand, user entity.User) *entity.EmojiCommand
	ParseArgs(arg string) (*entity.EmojiCommand, error)
	GenerateEmojiMessage(emojiMetaRows [][]entity.EmojiMeta, stickerSet *models.StickerSet, emojiArgs *entity.EmojiCommand) []entity.EmojiMeta
}

type Module struct {
	directories sync.Map
	logger      *slog.Logger
}

func NewProcessingModule(logger *slog.Logger) *Module {
	logger = logger.With(slog.String("module", "processing emoji-gen"))
	return &Module{
		directories: sync.Map{},
		logger:      logger,
	}
}
