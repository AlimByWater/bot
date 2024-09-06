package user_method

import (
	"arimadj-helper/internal/entity"
	"context"
)

type layoutUC interface {
	GetUserLayout(ctx context.Context, userID, initiatorUserID int) (entity.UserLayout, error)
	UpdateLayoutFull(ctx context.Context, userID, initiatorUserID int, layout entity.UserLayout) error
}
