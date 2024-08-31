func (r *LayoutRepository) LayoutByUserID(ctx context.Context, userID int) (entity.UserLayout, error) {
	query := `SELECT layout_id, name, background, layout, creator, editors FROM user_layouts WHERE creator = $1`
	var layout entity.UserLayout
	var editorsJSON string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&layout.LayoutID,
		&layout.Name,
		&layout.Background,
		&layout.Layout,
		&layout.Creator,
		&editorsJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.UserLayout{}, fmt.Errorf("layout not found for user %d", userID)
		}
		return entity.UserLayout{}, fmt.Errorf("error querying layout: %w", err)
	}

	// Parse editors JSON
	if err := json.Unmarshal([]byte(editorsJSON), &layout.Editors); err != nil {
		return entity.UserLayout{}, fmt.Errorf("error parsing editors: %w", err)
	}

	return layout, nil
}

func (r *LayoutRepository) LayoutByID(ctx context.Context, layoutID string) (entity.UserLayout, error) {
	query := `SELECT layout_id, name, background, layout, creator, editors FROM user_layouts WHERE layout_id = $1`
	var layout entity.UserLayout
	var editorsJSON string

	err := r.db.QueryRowContext(ctx, query, layoutID).Scan(
		&layout.LayoutID,
		&layout.Name,
		&layout.Background,
		&layout.Layout,
		&layout.Creator,
		&editorsJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.UserLayout{}, fmt.Errorf("layout not found with ID %s", layoutID)
		}
		return entity.UserLayout{}, fmt.Errorf("error querying layout: %w", err)
	}

	// Parse editors JSON
	if err := json.Unmarshal([]byte(editorsJSON), &layout.Editors); err != nil {
		return entity.UserLayout{}, fmt.Errorf("error parsing editors: %w", err)
	}

	return layout, nil
}

func (r *LayoutRepository) UpdateLayout(ctx context.Context, layout entity.UserLayout) error {
	query := `
		UPDATE user_layouts 
		SET name = $1, background = $2, layout = $3, editors = $4
		WHERE layout_id = $5
	`
	editorsJSON, err := json.Marshal(layout.Editors)
	if err != nil {
		return fmt.Errorf("error marshaling editors: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query, layout.Name, layout.Background, layout.Layout, editorsJSON, layout.LayoutID)
	if err != nil {
		return fmt.Errorf("error updating layout: %w", err)
	}

	return nil
}

func (r *LayoutRepository) LogLayoutChange(ctx context.Context, change entity.LayoutChange) error {
	query := `
		INSERT INTO layout_changes (layout_id, user_id, change_type, change_data, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, change.LayoutID, change.UserID, change.ChangeType, change.ChangeData, time.Now())
	if err != nil {
		return fmt.Errorf("error logging layout change: %w", err)
	}

	return nil
}

func (r *LayoutRepository) IsLayoutOwner(ctx context.Context, layoutID string, userID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_layouts WHERE layout_id = $1 AND creator = $2)`
	var isOwner bool

	err := r.db.QueryRowContext(ctx, query, layoutID, userID).Scan(&isOwner)
	if err != nil {
		return false, fmt.Errorf("error checking layout ownership: %w", err)
	}

	return isOwner, nil
}

func (r *LayoutRepository) AddLayoutEditor(ctx context.Context, layoutID string, editorID int) error {
	query := `
		UPDATE user_layouts 
		SET editors = array_append(editors, $2)
		WHERE layout_id = $1 AND NOT $2 = ANY(editors)
	`
	result, err := r.db.ExecContext(ctx, query, layoutID, editorID)
	if err != nil {
		return fmt.Errorf("error adding layout editor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("layout not found or editor already exists")
	}

	return nil
}

func (r *LayoutRepository) RemoveLayoutEditor(ctx context.Context, layoutID string, editorID int) error {
	query := `
		UPDATE user_layouts 
		SET editors = array_remove(editors, $2)
		WHERE layout_id = $1 AND $2 = ANY(editors)
	`
	result, err := r.db.ExecContext(ctx, query, layoutID, editorID)
	if err != nil {
		return fmt.Errorf("error removing layout editor: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("layout not found or editor doesn't exist")
	}

	return nil
}

func (r *LayoutRepository) GetDefaultLayout(ctx context.Context) (entity.UserLayout, error) {
	query := `SELECT layout_id, name, background, layout, creator, editors FROM user_layouts WHERE is_default = true LIMIT 1`
	var layout entity.UserLayout
	var editorsJSON string

	err := r.db.QueryRowContext(ctx, query).Scan(
		&layout.LayoutID,
		&layout.Name,
		&layout.Background,
		&layout.Layout,
		&layout.Creator,
		&editorsJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.UserLayout{}, errors.New("default layout not found")
		}
		return entity.UserLayout{}, fmt.Errorf("error querying default layout: %w", err)
	}

	// Parse editors JSON
	if err := json.Unmarshal([]byte(editorsJSON), &layout.Editors); err != nil {
		return entity.UserLayout{}, fmt.Errorf("error parsing editors: %w", err)
	}

	return layout, nil
}
