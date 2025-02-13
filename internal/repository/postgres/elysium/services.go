package elysium

import (
	"context"
	"database/sql"
	"elysium/internal/entity"
	"fmt"
)

const servicesTable = "services"

func (r *Repository) CreateService(ctx context.Context, service entity.Service) (entity.Service, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (bot_id, name, description, price, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, servicesTable)

	err := r.db.QueryRowxContext(ctx, query,
		service.BotID,
		service.Name,
		service.Description,
		service.Price,
		service.IsActive,
	).Scan(&service.ID, &service.CreatedAt)

	if err != nil {
		return entity.Service{}, fmt.Errorf("failed to create service: %w", err)
	}

	return service, nil
}

func (r *Repository) InitServices(ctx context.Context, services []entity.Service) error {
	for i, service := range services {
		query := fmt.Sprintf(`
			INSERT INTO %s (bot_id, name, description, price, is_active)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (bot_id, name) DO UPDATE SET
				description = EXCLUDED.description,                                                                                                                                                                           
				price = EXCLUDED.price,                                                                                                                                                                                       
				is_active = EXCLUDED.is_active,                                                                                                                                                                               
				updated_at = NOW()
			RETURNING id
		`, servicesTable)

		err := r.db.QueryRowContext(ctx, query,
			service.BotID,
			service.Name,
			service.Description,
			service.Price,
			service.IsActive,
		).Scan(&services[i].ID)
		if err != nil {
			return fmt.Errorf("failed to init service: %w", err)
		}
	}
	return nil
}

func (r *Repository) UpdateService(ctx context.Context, service entity.Service) (entity.Service, error) {
	query := fmt.Sprintf(`
		UPDATE %s SET
			name = $1,
			description = $2,
			price = $3,
			is_active = $4,
			updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`, servicesTable)

	var updatedAt sql.NullTime
	err := r.db.QueryRowxContext(ctx, query,
		service.Name,
		service.Description,
		service.Price,
		service.IsActive,
		service.ID,
	).Scan(&updatedAt)

	if err != nil {
		return entity.Service{}, fmt.Errorf("failed to update service: %w", err)
	}

	if updatedAt.Valid {
		service.UpdatedAt = updatedAt.Time
	}
	return service, nil
}

func (r *Repository) GetServiceByID(ctx context.Context, serviceID int) (entity.Service, error) {
	var service entity.Service
	query := fmt.Sprintf(`
		SELECT id, bot_id, name, description, price, is_active, created_at, updated_at
		FROM %s 
		WHERE id = $1
	`, servicesTable)

	err := r.db.GetContext(ctx, &service, query, serviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.Service{}, fmt.Errorf("service not found")
		}
		return entity.Service{}, fmt.Errorf("failed to get service: %w", err)
	}

	return service, nil
}

func (r *Repository) ListServicesByBotID(ctx context.Context, botID int64) ([]entity.Service, error) {
	var services []entity.Service
	query := fmt.Sprintf(`
		SELECT id, bot_id, name, description, price, is_active, created_at, updated_at
		FROM %s 
		WHERE bot_id = $1
		ORDER BY created_at DESC
	`, servicesTable)

	err := r.db.SelectContext(ctx, &services, query, botID)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	return services, nil
}

func (r *Repository) GetAllServices(ctx context.Context) ([]entity.Service, error) {
	var services []entity.Service
	query := fmt.Sprintf(`
		SELECT id, bot_id, name, description, price, is_active, created_at, updated_at
		FROM %s 
		ORDER BY bot_id, created_at DESC
	`, servicesTable)

	err := r.db.SelectContext(ctx, &services, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all services: %w", err)
	}

	return services, nil
}

func (r *Repository) DeleteServiceHard(ctx context.Context, serviceID int) error {
	query := fmt.Sprintf(`
		DELETE FROM %s 
		WHERE id = $1
	`, servicesTable)

	result, err := r.db.ExecContext(ctx, query, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service not found")
	}

	return nil
}
