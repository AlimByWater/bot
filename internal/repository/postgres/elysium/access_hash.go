package elysium

import (
	"context"
	"elysium/internal/entity"
	"fmt"
)

func (r *Repository) CreateOrUpdateAccessHash(ctx context.Context, accessHash entity.AccessHash) error {
	query := `
        INSERT INTO access_hashes (chat_id, username, hash, peer_id, created_at)
        VALUES ($1, $2, $3, $4, NOW())
        ON CONFLICT (chat_id) DO UPDATE 
        SET username = $2, hash = $3, peer_id = $4`

	_, err := r.db.ExecContext(ctx, query,
		accessHash.ChatID,
		accessHash.Username,
		accessHash.Hash,
		accessHash.PeerID)
	if err != nil {
		return fmt.Errorf("failed to create/update access hash: %w", err)
	}

	return nil
}

func (r *Repository) GetAccessHash(ctx context.Context, chatID string) (entity.AccessHash, error) {
	query := `SELECT chat_id, username, hash, peer_id, created_at FROM access_hashes WHERE chat_id = $1`

	var accessHash entity.AccessHash
	err := r.db.GetContext(ctx, &accessHash, query, chatID)
	if err != nil {
		return entity.AccessHash{}, fmt.Errorf("failed to get access hash: %w", err)
	}

	return accessHash, nil
}

func (r *Repository) GetAllAccessHashes(ctx context.Context) ([]entity.AccessHash, error) {
	query := `SELECT chat_id, username, hash, peer_id, created_at FROM access_hashes`

	var accessHashes []entity.AccessHash
	err := r.db.SelectContext(ctx, &accessHashes, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all access hashes: %w", err)
	}

	return accessHashes, nil
}

func (r *Repository) DeleteAccessHash(ctx context.Context, chatID string) error {
	query := `DELETE FROM access_hashes WHERE chat_id = $1`

	result, err := r.db.ExecContext(ctx, query, chatID)
	if err != nil {
		return fmt.Errorf("failed to delete access hash: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("access hash not found for chat_id: %s", chatID)
	}

	return nil
}
