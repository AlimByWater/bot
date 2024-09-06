package layout_methods

import (
	"arimadj-helper/internal/entity"
	"context"
)

type layoutUC interface {
	GetUserLayout(ctx context.Context, userID, initiatorUserID int) (entity.UserLayout, error)
	GetLayout(ctx context.Context, layoutID int, initiatorUserID int) (entity.UserLayout, error)
	GetLayoutByName(ctx context.Context, layoutName string, initiatorUserID int) (entity.UserLayout, error)
	UpdateLayoutFull(ctx context.Context, layoutID int, initiatorUserID int, layout entity.UserLayout) error
	AddLayoutEditor(ctx context.Context, layoutID int, initiatorUserID, editorID int) error
	RemoveLayoutEditor(ctx context.Context, layoutID int, initiatorUserID, editorID int) error
}
