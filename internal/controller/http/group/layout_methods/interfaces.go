package layout_methods

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
)

type layoutUC interface {
	GetUserLayout(ctx context.Context, userID, initiatorUserID int) (entity.UserLayout, error)
	GetLayout(ctx context.Context, layoutID string, initiatorUserID int) (entity.UserLayout, error)
	UpdateLayoutFull(ctx context.Context, layoutID string, initiatorUserID int, layout entity.UserLayout) error
	AddLayoutEditor(ctx context.Context, layoutID string, initiatorUserID, editorID int) error
	RemoveLayoutEditor(ctx context.Context, layoutID string, initiatorUserID, editorID int) error
}

func getUserID(c *gin.Context) (int, error) {
	u, exist := c.Get("userId")
	if !exist {
		return 0, fmt.Errorf("user id not found in token")
	}
	userIdString, ok := u.(string)
	if !ok {
		return 0, fmt.Errorf("user id is not a string")
	}

	userId, err := strconv.Atoi(userIdString)
	if err != nil {
		return 0, fmt.Errorf("user id is not a number: %w", err)
	}

	return userId, nil
}
