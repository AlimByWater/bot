package elysium

import (
	"context"
	"database/sql"
	"elysium/internal/entity"
	"errors"
	"fmt"
)

func (r *Repository) GetDefaultLayout(ctx context.Context) (entity.UserLayout, error) {
	query := `
    SELECT id, name, creator_id, stream_url, background, created_at, updated_at
    FROM elysium.user_layouts
    WHERE name = 'default_layout_1'
    LIMIT 1
    `
	var layout entity.UserLayout
	err := r.db.QueryRowContext(ctx, query).Scan(
		&layout.ID,
		&layout.Name,
		&layout.Creator,
		&layout.StreamURL,
		&layout.Background,
		&layout.CreatedAt,
		&layout.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserLayout{}, entity.ErrLayoutNotFound
		}
		return entity.UserLayout{}, fmt.Errorf("query layout by ID: %w", err)
	}

	elements, err := r.ElementsByLayoutID(ctx, layout.ID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("elements by layout: %w", err)
	}

	layout.Elements = elements

	return layout, nil
}

func (r *Repository) LayoutByName(ctx context.Context, layoutName string) (entity.UserLayout, error) {
	query := `
    SELECT id, name, creator_id, stream_url, background, created_at, updated_at
    FROM elysium.user_layouts
    WHERE name = $1
    `
	var layout entity.UserLayout
	err := r.db.QueryRowContext(ctx, query, layoutName).Scan(
		&layout.ID,
		&layout.Name,
		&layout.Creator,
		&layout.StreamURL,
		&layout.Background,
		&layout.CreatedAt,
		&layout.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserLayout{}, entity.ErrLayoutNotFound
		}
		return entity.UserLayout{}, fmt.Errorf("query layout by ID: %w", err)
	}

	editors, err := r.EditorByLayoutID(ctx, layout.ID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("editors by layout: %w", err)
	}

	layout.Editors = editors

	elements, err := r.ElementsByLayoutID(ctx, layout.ID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("elements by layout: %w", err)
	}

	layout.Elements = elements

	return layout, nil
}

// LayoutByUserID получает макет пользователя по его ID
func (r *Repository) LayoutByUserID(ctx context.Context, userID int) (entity.UserLayout, error) {
	query := `
    SELECT id, name, creator_id, stream_url, background, created_at, updated_at
    FROM elysium.user_layouts
    WHERE creator_id = $1
    LIMIT 1
    `
	var layout entity.UserLayout
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&layout.ID,
		&layout.Name,
		&layout.Creator,
		&layout.StreamURL,
		&layout.Background,
		&layout.CreatedAt,
		&layout.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserLayout{}, entity.ErrLayoutNotFound
		}
		return entity.UserLayout{}, fmt.Errorf("query layout by user ID: %w", err)
	}

	editors, err := r.EditorByLayoutID(ctx, layout.ID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("editors by layout: %w", err)
	}

	layout.Editors = editors

	elements, err := r.ElementsByLayoutID(ctx, layout.ID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("elements by layout: %w", err)
	}

	layout.Elements = elements

	return layout, nil
}

// LayoutByID получает макет по его ID
func (r *Repository) LayoutByID(ctx context.Context, layoutID int) (entity.UserLayout, error) {
	query := `
    SELECT id, name, creator_id, stream_url, background, created_at, updated_at
    FROM elysium.user_layouts
    WHERE id = $1
    `
	var layout entity.UserLayout
	err := r.db.QueryRowContext(ctx, query, layoutID).Scan(
		&layout.ID,
		&layout.Name,
		&layout.Creator,
		&layout.StreamURL,
		&layout.Background,
		&layout.CreatedAt,
		&layout.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.UserLayout{}, entity.ErrLayoutNotFound
		}
		return entity.UserLayout{}, fmt.Errorf("query layout by ID: %w", err)
	} // Получаем редакторов макета

	editors, err := r.EditorByLayoutID(ctx, layoutID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("editors by layout: %w", err)
	}

	layout.Editors = editors

	elements, err := r.ElementsByLayoutID(ctx, layoutID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("elements by layout: %w", err)
	}

	layout.Elements = elements

	return layout, nil
}

func (r *Repository) ElementsByLayoutID(ctx context.Context, layoutID int) ([]entity.LayoutElement, error) {
	var elements []entity.LayoutElement
	// Получаем элементы макета
	elementsQuery := `
				SELECT el.id, el.properties, el.position_x, el.position_y, el.position_z, el.width, el.height, el.is_public, el.is_removable,
				       re.id, re.name, re.type, re.default_properties, re.description, re.is_public, re.is_paid
				FROM elysium.layout_elements el
				JOIN elysium.root_elements re ON re.id = root_element_id
				WHERE layout_id = $1
				`
	elemRows, err := r.db.QueryContext(ctx, elementsQuery, layoutID)
	if err != nil {
		return nil, fmt.Errorf("query layout elements: %w", err)
	}
	defer elemRows.Close()

	for elemRows.Next() {
		var elem entity.LayoutElement
		if err := elemRows.Scan(
			&elem.ID,
			&elem.Properties,
			&elem.Position.X,
			&elem.Position.Y,
			&elem.Position.Z,
			&elem.Position.Width,
			&elem.Position.Height,
			&elem.IsPublic,
			&elem.IsRemovable,
			&elem.RootElement.ID,
			&elem.RootElement.Name,
			&elem.RootElement.Type,
			&elem.RootElement.DefaultProperties,
			&elem.RootElement.Description,
			&elem.RootElement.IsPublic,
			&elem.RootElement.IsPaid,
		); err != nil {
			return nil, fmt.Errorf("scan layout element: %w", err)
		}

		elements = append(elements, elem)

	}

	return elements, nil
}

// EditorByLayoutID получает список редакторов макета по его ID
func (r *Repository) EditorByLayoutID(ctx context.Context, layoutID int) ([]int, error) {
	query := `
 SELECT editor_id
 FROM elysium.layout_editors
 WHERE layout_id = $1
 `
	rows, err := r.db.QueryContext(ctx, query, layoutID)
	if err != nil {
		return nil, fmt.Errorf("query layout editors: %w", err)
	}
	defer rows.Close()

	var editors []int
	for rows.Next() {
		var editorID int
		if err := rows.Scan(&editorID); err != nil {
			return nil, fmt.Errorf("scan editor ID: %w", err)
		}
		editors = append(editors, editorID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate over rows: %w", err)
	}

	return editors, nil
}

func (r *Repository) UpdateLayoutFull(ctx context.Context, layout entity.UserLayout) error {
	err := r.execTX(ctx, func(q *queries) error {
		updateQuery := `
			UPDATE elysium.user_layouts
			SET name = $1, background = $2, stream_url = $3, updated_at = CURRENT_TIMESTAMP
			WHERE id = $4
			`
		_, err := q.db.ExecContext(ctx, updateQuery, layout.Name, layout.Background, layout.StreamURL, layout.ID)
		if err != nil {
			return fmt.Errorf("update layout: %w", err)
		}

		// Удаляем существующие элементы макета
		deleteElementsQuery := `DELETE FROM elysium.layout_elements WHERE layout_id = $1`
		_, err = q.db.ExecContext(ctx, deleteElementsQuery, layout.ID)
		if err != nil {
			return fmt.Errorf("delete existing layout elements: %w", err)
		}

		// Добавляем новые элементы макета
		insertElementQuery := `
			INSERT INTO elysium.layout_elements (layout_id, root_element_id, properties, position_x, position_y, position_z, width, height, is_public, is_removable)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
			`
		for _, elem := range layout.Elements {
			err = q.db.QueryRowContext(ctx, insertElementQuery, layout.ID, elem.RootElement.ID, elem.Properties, elem.Position.X, elem.Position.Y, elem.Position.Z, elem.Position.Width, elem.Position.Height, elem.IsPublic, elem.IsRemovable).Scan(&elem.ID)
			if err != nil {
				return fmt.Errorf("insert layout element: %w", err)
			}
		}

		// Обновляем список редакторов
		deleteEditorsQuery := `DELETE FROM elysium.layout_editors WHERE layout_id = $1`
		_, err = q.db.ExecContext(ctx, deleteEditorsQuery, layout.ID)
		if err != nil {
			return fmt.Errorf("delete existing layout editors: %w", err)
		}

		insertEditorQuery := `INSERT INTO elysium.layout_editors (layout_id, editor_id) VALUES ($1, $2)`
		for _, editorID := range layout.Editors {
			_, err = q.db.ExecContext(ctx, insertEditorQuery, layout.ID, editorID)
			if err != nil {
				return fmt.Errorf("insert layout editor: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("exec tx: %w", err)
	}
	return nil

}

// LogLayoutChange логирует изменение макета
func (r *Repository) LogLayoutChange(ctx context.Context, change entity.LayoutChange) error {
	query := `
    INSERT INTO elysium.layout_changes (layout_id, user_id, change_type, details, timestamp)
    VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
    `
	_, err := r.db.ExecContext(ctx, query, change.LayoutID, change.UserID, change.ChangeType, change.Details)
	if err != nil {
		return fmt.Errorf("log layout change: %w", err)
	}
	return nil
}

// IsLayoutOwner проверяет, является ли пользователь владельцем макета
func (r *Repository) IsLayoutOwner(ctx context.Context, layoutID int, userID int) (bool, error) {
	query := `
    SELECT EXISTS(SELECT 1 FROM elysium.user_layouts WHERE id = $1 AND creator_id = $2)
    `
	var isOwner bool
	err := r.db.QueryRowContext(ctx, query, layoutID, userID).Scan(&isOwner)
	if err != nil {
		return false, fmt.Errorf("check layout ownership: %w", err)
	}
	return isOwner, nil
}

// AddLayoutEditor добавляет редактора к макету
func (r *Repository) AddLayoutEditor(ctx context.Context, layoutID int, editorID int) error {
	query := `
    INSERT INTO elysium.layout_editors (layout_id, editor_id)
    VALUES ($1, $2)
    ON CONFLICT (layout_id, editor_id) DO NOTHING
    `
	_, err := r.db.ExecContext(ctx, query, layoutID, editorID)
	if err != nil {
		return fmt.Errorf("add layout editor: %w", err)
	}
	return nil
}

// RemoveLayoutEditor удаляет редактора из макета
func (r *Repository) RemoveLayoutEditor(ctx context.Context, layoutID int, editorID int) error {
	query := `
    DELETE FROM elysium.layout_editors
    WHERE layout_id = $1 AND editor_id = $2
    `
	_, err := r.db.ExecContext(ctx, query, layoutID, editorID)
	if err != nil {
		return fmt.Errorf("remove layout editor: %w", err)
	}
	return nil
}

// CreateLayout создает новый макет для пользователя
func (r *Repository) CreateLayout(ctx context.Context, layout entity.UserLayout) error {
	err := r.execTX(ctx, func(q *queries) error {
		insertLayoutQuery := `
		INSERT INTO elysium.user_layouts (name, creator_id, stream_url, background)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
		`
		err := q.db.QueryRowContext(ctx, insertLayoutQuery, layout.Name, layout.Creator, layout.StreamURL, layout.Background).
			Scan(&layout.ID, &layout.CreatedAt, &layout.UpdatedAt)
		if err != nil {
			return fmt.Errorf("insert layout: %w", err)
		}

		// Добавляем новые элементы макета
		insertElementQuery := `
			INSERT INTO elysium.layout_elements (layout_id, root_element_id, properties, position_x, position_y, position_z, width, height, is_public, is_removable)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
			`
		for _, elem := range layout.Elements {
			err = q.db.QueryRowContext(ctx, insertElementQuery, layout.ID, elem.RootElement.ID, elem.Properties, elem.Position.X, elem.Position.Y, elem.Position.Z, elem.Position.Width, elem.Position.Height, elem.IsPublic, elem.IsRemovable).Scan(&elem.ID)
			if err != nil {
				return fmt.Errorf("insert layout element: %w", err)
			}
		}

		// Добавляем новые редакторы макета
		insertEditorQuery := `
		INSERT INTO elysium.layout_editors (layout_id, editor_id)
		VALUES ($1, $2)
		`
		for _, editorID := range layout.Editors {
			_, err = q.db.ExecContext(ctx, insertEditorQuery, layout.ID, editorID)
			if err != nil {
				return fmt.Errorf("insert layout editor: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("exec tx: %w", err)
	}
	return nil
}

// DeleteLayout удаляет макет по его ID
func (r *Repository) DeleteLayout(ctx context.Context, layoutID int) error {
	err := r.execTX(ctx, func(q *queries) error {
		// Удаляем элементы макета
		deleteElementsQuery := `DELETE FROM elysium.layout_elements WHERE layout_id = $1`
		_, err := q.db.ExecContext(ctx, deleteElementsQuery, layoutID)
		if err != nil {
			return fmt.Errorf("delete layout elements: %w", err)
		}

		// Удаляем редакторов макета
		deleteEditorsQuery := `DELETE FROM elysium.layout_editors WHERE layout_id = $1`
		_, err = q.db.ExecContext(ctx, deleteEditorsQuery, layoutID)
		if err != nil {
			return fmt.Errorf("delete layout editors: %w", err)
		}

		// Удаляем сам макет
		deleteLayoutQuery := `DELETE FROM elysium.user_layouts WHERE id = $1`
		result, err := q.db.ExecContext(ctx, deleteLayoutQuery, layoutID)
		if err != nil {
			return fmt.Errorf("delete layout: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return entity.ErrLayoutNotFound
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("exec tx: %w", err)
	}
	return nil
}
