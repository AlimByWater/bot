package elysium

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"strconv"
)

func (r *Repository) GetEmojiPackByPackLink(ctx context.Context, packLink string) (entity.EmojiPack, error) {
	query := `
		SELECT
			id,
			pack_link,
			creator_telegram_id,
			pack_title,
			telegram_file_id,
			initial_command,
			bot_id as "bot_id",
			emoji_count,
			deleted,
			created_at,
			updated_at
		FROM emoji_packs
		WHERE pack_link = $1
	`

	var pack entity.EmojiPack
	err := r.db.GetContext(ctx, &pack, query, packLink)
	if err != nil {
		return entity.EmojiPack{}, fmt.Errorf("failed to get emoji pack by pack link: %w", err)
	}

	return pack, nil
}

func (r *Repository) GetEmojiPacksByCreator(ctx context.Context, creator int64, deleted bool) ([]entity.EmojiPack, error) {
	query := `
		SELECT
			id,
			pack_link,
			creator_telegram_id,
			pack_title,
			telegram_file_id,
			initial_command,
			bot_id as "bot_id",
			emoji_count,
			deleted,
			created_at,
			updated_at
		FROM emoji_packs
		WHERE creator_telegram_id = $1 AND deleted = $2
	`

	var packs []entity.EmojiPack
	err := r.db.SelectContext(ctx, &packs, query, creator, deleted)
	if err != nil {
		return nil, fmt.Errorf("failed to get emoji packs by creator: %w", err)
	}

	return packs, nil
}

func (r *Repository) SetEmojiPackDeleted(ctx context.Context, packName string) error {
	query := `
		UPDATE emoji_packs
		SET deleted = true
		WHERE pack_link = $1
	`

	_, err := r.db.ExecContext(ctx, query, packName)
	if err != nil {
		return fmt.Errorf("failed to set emoji pack as deleted: %w", err)
	}

	return nil
}

func (r *Repository) UnsetEmojiPackDeleted(ctx context.Context, packName string) error {
	query := `
		UPDATE emoji_packs
		SET deleted = false
		WHERE pack_link = $1
	`

	_, err := r.db.ExecContext(ctx, query, packName)
	if err != nil {
		return fmt.Errorf("failed to unset emoji pack as deleted: %w", err)
	}

	return nil
}

func (r *Repository) CreateNewEmojiPack(ctx context.Context, pack entity.EmojiPack) (entity.EmojiPack, error) {
	query := `
		INSERT INTO emoji_packs (
			pack_link,
			creator_telegram_id,
			pack_title,
			telegram_file_id,
			initial_command,
			bot_id,
			emoji_count,
		
			deleted,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	if pack.BotID > 0 {
		var err error
		pack.BotID, err = strconv.ParseInt(fmt.Sprintf("-100%d", pack.BotID), 10, 64)
		if err != nil {
			return entity.EmojiPack{}, fmt.Errorf("failed to parse bot id: %w", err)
		}
	}

	err := r.db.QueryRowContext(ctx, query,
		pack.PackLink,
		pack.CreatorTelegramID,
		pack.PackTitle,
		pack.TelegramFileID,
		pack.InitialCommand,
		pack.BotID,
		pack.EmojiCount,
		pack.Deleted,
		pack.CreatedAt,
		pack.UpdatedAt,
	).Scan(&pack.ID)
	if err != nil {
		return entity.EmojiPack{}, fmt.Errorf("failed to create new emoji pack: %w", err)
	}

	return pack, nil
}

func (r *Repository) UpdateEmojiCount(ctx context.Context, packID int64, emojiCount int) error {
	query := `
		UPDATE emoji_packs
		SET emoji_count = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, emojiCount, packID)
	if err != nil {
		return fmt.Errorf("failed to update emoji count: %w", err)
	}

	return nil
}
