package layout

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"time"
)

type LayoutChange struct {
	UserID    int
	LayoutID  string
	Timestamp time.Time
	Action    string
	Details   string
}

func (m *Module) GetUserLayout(ctx context.Context, userID, initiatorUserID int) (entity.UserLayout, error) {
	layout, err := m.repo.LayoutByUserID(ctx, userID)
	if err != nil {
		return entity.UserLayout{}, fmt.Errorf("failed to get user layout: %w", err)
	}

	if !m.hasViewPermission(layout, initiatorUserID) {
		return entity.UserLayout{}, entity.ErrNoPermission
	}

	return layout, nil
}

func (m *Module) UpdateLayoutFull(ctx context.Context, userID, initiatorUserID int, updatedLayout entity.UserLayout) error {
	currentLayout, err := m.repo.LayoutByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get current layout: %w", err)
	}

	if !m.hasEditPermission(currentLayout, initiatorUserID) {
		return entity.ErrNoPermission
	}

	updatedLayout.Creator = currentLayout.Creator
	updatedLayout.Editors = currentLayout.Editors

	err = m.repo.UpdateLayout(ctx, updatedLayout)
	if err != nil {
		return fmt.Errorf("failed to update layout: %w", err)
	}

	m.logLayoutChange(ctx, initiatorUserID, updatedLayout.LayoutID, "UpdateLayoutFull", fmt.Sprintf("Layout updated for user %d", userID))

	return nil
}

func (m *Module) AddEditor(ctx context.Context, layoutID string, editorID int) error {
	layout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return fmt.Errorf("failed to get layout: %w", err)
	}

	layout.Editors = append(layout.Editors, editorID)

	err = m.repo.UpdateLayout(ctx, layout)
	if err != nil {
		return fmt.Errorf("failed to update layout with new editor: %w", err)
	}

	m.logLayoutChange(ctx, layout.Creator, layoutID, "AddEditor", fmt.Sprintf("Added editor %d to layout", editorID))

	return nil
}

func (m *Module) RemoveEditor(ctx context.Context, layoutID string, editorID int) error {
	layout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return fmt.Errorf("failed to get layout: %w", err)
	}

	for i, editor := range layout.Editors {
		if editor == editorID {
			layout.Editors = append(layout.Editors[:i], layout.Editors[i+1:]...)
			break
		}
	}

	err = m.repo.UpdateLayout(ctx, layout)
	if err != nil {
		return fmt.Errorf("failed to update layout after removing editor: %w", err)
	}

	m.logLayoutChange(ctx, layout.Creator, layoutID, "RemoveEditor", fmt.Sprintf("Removed editor %d from layout", editorID))

	return nil
}

func (m *Module) IsEditor(ctx context.Context, layoutID string, userID int) (bool, error) {
	layout, err := m.repo.LayoutByID(ctx, layoutID)
	if err != nil {
		return false, fmt.Errorf("failed to get layout: %w", err)
	}

	return m.hasEditPermission(layout, userID), nil
}

func (m *Module) hasEditPermission(layout entity.UserLayout, userID int) bool {
	if layout.Creator == userID {
		return true
	}
	for _, editor := range layout.Editors {
		if editor == userID {
			return true
		}
	}
	return false
}

func (m *Module) hasViewPermission(layout entity.UserLayout, userID int) bool {
	return layout.Creator == userID || m.hasEditPermission(layout, userID)
}

func (m *Module) logLayoutChange(ctx context.Context, userID int, layoutID string, action, details string) {
	change := LayoutChange{
		UserID:    userID,
		LayoutID:  layoutID,
		Timestamp: time.Now(),
		Action:    action,
		Details:   details,
	}
	m.logger.Info("Layout change", 
		"userID", change.UserID,
		"layoutID", change.LayoutID,
		"action", change.Action,
		"details", change.Details,
		"timestamp", change.Timestamp)
}
