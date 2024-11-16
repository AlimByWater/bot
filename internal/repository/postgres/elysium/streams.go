package elysium

import (
	"context"
	"elysium/internal/entity"
	"fmt"
)

func (r *Repository) AvailableStreams(ctx context.Context) ([]*entity.Stream, error) {
	query := `
		SELECT slug, name, link, logo_link, icon_link, on_click_link, priority FROM elysium.streams
WHERE enabled = true
`
	var streams []*entity.Stream
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query streams: %w", err)
	}

	for rows.Next() {
		var stream entity.Stream
		err := rows.Scan(
			&stream.Slug,
			&stream.Name,
			&stream.Link,
			&stream.LogoLink,
			&stream.IconLink,
			&stream.OnClickLink,
			&stream.Priority,
		)
		if err != nil {
			return nil, fmt.Errorf("scan stream: %w", err)
		}
		streams = append(streams, &stream)
	}

	return streams, nil
}
