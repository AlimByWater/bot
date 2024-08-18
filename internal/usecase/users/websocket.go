package users

import "arimadj-helper/internal/entity"

func (m *Module) WebsocketInfo() (entity.WebsocketInfo, error) {
	count := m.GetOnlineUsersCount()

	return entity.WebsocketInfo{
		OnlineUsersCount: count,
	}, nil
}
