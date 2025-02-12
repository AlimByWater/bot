package elysium

import (
	"context"
	"elysium/internal/entity"
	"fmt"
)

func (r *Repository) CreatePromoCode(ctx context.Context, promo entity.PromoCode) (entity.PromoCode, error) {
	query := `
		INSERT INTO promo_codes (
			code, type, user_id, bonus_redeemer, bonus_referrer, usage_limit, usage_count, valid_from, valid_to, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (code) DO NOTHING
		RETURNING code, type, user_id, bonus_redeemer, bonus_referrer, usage_limit, usage_count, valid_from, valid_to, created_at, updated_at
	`
	var result entity.PromoCode
	err := r.db.QueryRowContext(ctx, query,
		promo.Code,
		promo.Type,
		promo.UserID,
		promo.BonusRedeemer,
		promo.BonusReferrer,
		promo.UsageLimit,
		promo.UsageCount,
		promo.ValidFrom,
		promo.ValidTo,
	).Scan(
		&result.Code,
		&result.Type,
		&result.UserID,
		&result.BonusRedeemer,
		&result.BonusReferrer,
		&result.UsageLimit,
		&result.UsageCount,
		&result.ValidFrom,
		&result.ValidTo,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return entity.PromoCode{}, fmt.Errorf("failed to insert promo code: %w", err)
	}
	return result, nil
}
