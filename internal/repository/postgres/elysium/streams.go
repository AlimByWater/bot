package elysium

import (
	"context"
	"elysium/internal/entity"
	"fmt"
)

func (r *Repository) AvailableStreams(ctx context.Context) ([]*entity.Stream, error) {
	query := `
		SELECT slug, link, logo_link FROM elysium.streams
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
			&stream.Link,
			&stream.LogoLink,
		)
		if err != nil {
			return nil, fmt.Errorf("scan stream: %w", err)
		}
		streams = append(streams, &stream)
	}

	return streams, nil
}
