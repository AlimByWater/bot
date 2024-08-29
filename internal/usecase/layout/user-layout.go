package layout

import (
	"arimadj-helper/internal/entity"
	"context"
)

func (m *Module) GetUserLayout(ctx context.Context, userID int) (entity.UserLayout, error) {
	return entity.UserLayout{}, nil
}
