package elysium

import (
	"arimadj-helper/internal/entity"
	"context"
)

func (r *Repository) LayoutByUserID(ctx context.Context, userID int) (entity.UserLayout, error) {
	query := `
		SELECT user_id, layout_id, is_public, background, layout, creator, editors
		FROM user_layouts
		WHERE user_id = $1
	`
	var layout entity.UserLayout
	var backgroundJSON, layoutJSON, editorsJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&layout.UserID,
		&layout.LayoutID,
		&layout.IsPublic,
		&backgroundJSON,
		&layoutJSON,
		&layout.Creator,
		&editorsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return entity.UserLayout{}, fmt.Errorf("layout not found for user ID %d", userID)
		}
		return entity.UserLayout{}, fmt.Errorf("error fetching layout: %w", err)
	}

	err = json.Unmarshal(backgroundJSON, &layout.Background)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("error unmarshaling background: %w", err)
	}

	err = json.Unmarshal(layoutJSON, &layout.Layout)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("error unmarshaling layout: %w", err)
	}

	err = json.Unmarshal(editorsJSON, &layout.Editors)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("error unmarshaling editors: %w", err)
	}

	return layout, nil
}
