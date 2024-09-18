package user_method

import (
	"context"
	"elysium/internal/entity"
)

type layoutUC interface {
	GetUserLayout(ctx context.Context, userID, initiatorUserID int) (entity.UserLayout, error)
	UpdateLayoutFull(ctx context.Context, userID, initiatorUserID int, layout entity.UserLayout) error
}
