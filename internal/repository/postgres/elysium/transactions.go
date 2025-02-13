package elysium

import (
	"context"
	"database/sql"
	"elysium/internal/entity"
	"encoding/json"
	"fmt"
	"time"
)

func (r *Repository) GetTransactionsByUserID(ctx context.Context, userID int) ([]entity.UserTransaction, error) {
	query := `
		SELECT 
			t.id,
			t.user_id,
			t.type,
			t.amount,
			t.status,
			t.provider,
			t.external_id,
			t.service_id,
			t.bot_id,
			t.description,
			t.promo_code_id,
			t.created_at,
			t.updated_at
		FROM user_transactions t
		LEFT JOIN services s ON t.service_id = s.id
		WHERE t.user_id = $1
		ORDER BY t.created_at DESC, t.id DESC
	`

	var transactions []entity.UserTransaction
	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var txn entity.UserTransaction
		var serviceID sql.NullInt64
		var botID sql.NullInt64
		var provider, externalID, promoCodeID sql.NullString

		err := rows.Scan(
			&txn.ID,
			&txn.UserID,
			&txn.Type,
			&txn.Amount,
			&txn.Status,
			&provider,
			&externalID,
			&serviceID,
			&botID,
			&txn.Description,
			&promoCodeID,
			&txn.CreatedAt,
			&txn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		if serviceID.Valid {
			txn.ServiceID = int(serviceID.Int64)
		}
		if botID.Valid {
			txn.BotID = botID.Int64
		}
		if provider.Valid {
			txn.Provider = provider.String
		}
		if externalID.Valid {
			txn.ExternalID = externalID.String
		}
		if promoCodeID.Valid {
			if promo, err := r.GetPromoCodeByID(ctx, promoCodeID.String); err == nil {
				txn.PromoCode = &promo
			}
		}

		transactions = append(transactions, txn)
	}

	return transactions, nil
}

func (r *Repository) GetTransactionByID(ctx context.Context, txnID string) (entity.UserTransaction, error) {
	query := `
		SELECT
			t.id,
			t.user_id,
			t.type,
			t.amount,
			t.status,
			t.provider,
			t.external_id,
			t.service_id,
			t.bot_id,
			t.description,
			t.promo_code_id,
			t.created_at,
			t.updated_at
		FROM user_transactions t
		LEFT JOIN services s ON t.service_id = s.id
		WHERE t.id = $1
	`

	var txn entity.UserTransaction
	var serviceID sql.NullInt64
	var botID sql.NullInt64
	var provider, externalID, description, promoCodeID sql.NullString

	err := r.db.QueryRowContext(ctx, query, txnID).Scan(
		&txn.ID,
		&txn.UserID,
		&txn.Type,
		&txn.Amount,
		&txn.Status,
		&provider,
		&externalID,
		&serviceID,
		&botID,
		&description,
		&promoCodeID,
		&txn.CreatedAt,
		&txn.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.UserTransaction{}, fmt.Errorf("transaction not found: %w", err)
		}
		return entity.UserTransaction{}, fmt.Errorf("failed to get transaction: %w", err)
	}

	if serviceID.Valid {
		txn.ServiceID = int(serviceID.Int64)
	}
	if botID.Valid {
		txn.BotID = botID.Int64
	}
	if provider.Valid {
		txn.Provider = provider.String
	}
	if externalID.Valid {
		txn.ExternalID = externalID.String
	}
	if description.Valid {
		txn.Description = description.String
	}
	if promoCodeID.Valid {
		if promo, err := r.GetPromoCodeByID(ctx, promoCodeID.String); err == nil {
			txn.PromoCode = &promo
		}
	}

	return txn, nil
}

func (r *Repository) ProcessTransaction(ctx context.Context, txnID string, userID int, balanceChange int, newStatus string) error {
	err := r.execTX(ctx, func(q *queries) error {
		// Обновляем статус транзакции
		if err := q.updateTransactionStatus(ctx, txnID, newStatus); err != nil {
			return err
		}
		// Обновляем баланс пользователя
		if err := q.updateUserBalance(ctx, userID, balanceChange); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	go func() {
		// Получаем обновленную транзакцию для аудита
		txn, err := r.GetTransactionByID(context.Background(), txnID)
		if err != nil {
			//r.logger.Error("failed to get transaction for audit log", "error", err.Error(), "txn_id", txnID)
			return // не блокируем основную операцию
		}

		// Формируем JSON-событие аудита
		auditEvent := entity.NewTransactionAuditEventFromTransaction(txn)
		payload, errMarshal := json.Marshal(auditEvent)
		if errMarshal != nil {
			//r.logger.Error("failed to marshal audit event in ProcessTransaction", "error", errMarshal.Error())
		} else {
			if err := r.SaveAuditLog(txn.ID, string(payload), auditEvent.Version, time.Now()); err != nil {
				fmt.Println("err", err.Error())
			}
		}
	}()

	return nil
}

func (r *Repository) GetPromoCodeByID(ctx context.Context, codeID string) (entity.PromoCode, error) {
	query := `
		SELECT
			code,
			type,
			user_id,
			bonus_redeemer,
			bonus_referrer,
			usage_limit,
			usage_count,
			valid_from,
			valid_to,
			created_at,
			updated_at
		FROM promo_codes
		WHERE code = $1
	`

	var promoCode entity.PromoCode
	var userID sql.NullInt64
	var validTo sql.NullTime

	err := r.db.QueryRowContext(ctx, query, codeID).Scan(
		&promoCode.Code,
		&promoCode.Type,
		&userID,
		&promoCode.BonusRedeemer,
		&promoCode.BonusReferrer,
		&promoCode.UsageLimit,
		&promoCode.UsageCount,
		&promoCode.ValidFrom,
		&validTo,
		&promoCode.CreatedAt,
		&promoCode.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return entity.PromoCode{}, fmt.Errorf("promo code not found: %w", err)
		}
		return entity.PromoCode{}, fmt.Errorf("failed to get promo code: %w", err)
	}

	if userID.Valid {
		promoCode.UserID = new(int)
		*promoCode.UserID = int(userID.Int64)
	}

	if validTo.Valid {
		promoCode.ValidTo = &validTo.Time
	}

	return promoCode, nil
}

// CreateTransaction создает транзакцию, используя атомарные операции с БД.
func (r *Repository) CreateTransaction(ctx context.Context, txn entity.UserTransaction) (entity.UserTransaction, error) {
	var resultTxn entity.UserTransaction

	err := r.execTX(ctx, func(q *queries) error {
		// 1. Блокируем баланс пользователя
		_, err := q.lockUserBalance(ctx, txn.UserID)
		if err != nil {
			return fmt.Errorf("failed to lock user balance: %w", err)
		}

		// Обновляем баланс с резервированием, если транзакция в статусе Pending.
		if txn.Status == entity.TransactionStatusPending {
			// Если транзакция списывает деньги (например, Withdrawal) – резервируем, вычитая сумму.
			if txn.Type == entity.TransactionTypeWithdrawal {
				var balanceChange int
				if txn.Amount > 0 {
					balanceChange = -txn.Amount
				} else {
					balanceChange = txn.Amount
				}
				if err := q.updateUserBalance(ctx, txn.UserID, balanceChange); err != nil {
					return fmt.Errorf("failed to reserve funds: %w", err)
				}
			}
			// Для остальных типов в состоянии Pending баланс не меняем.
		} else {
			// Если транзакция создаётся не в статусе Pending – обновляем баланс стандартно.
			var balanceChange int
			switch txn.Type {
			case entity.TransactionTypeDeposit,
				entity.TransactionTypeRefund,
				entity.TransactionTypePromoRedeem,
				entity.TransactionTypeReferralBonus,
				entity.TransactionTypePromoTransfer:
				balanceChange = txn.Amount
			case entity.TransactionTypeWithdrawal:
				if txn.Amount > 0 {
					balanceChange = -txn.Amount
				} else {
					balanceChange = txn.Amount
				}
			}
			if err := q.updateUserBalance(ctx, txn.UserID, balanceChange); err != nil {
				return fmt.Errorf("failed to update user balance: %w", err)
			}
		}

		// 3. Вставляем транзакцию
		var err2 error
		resultTxn, err2 = q.insertTransaction(ctx, txn)
		if err2 != nil {
			return fmt.Errorf("failed to insert transaction: %w", err2)
		}

		return nil
	})

	if err != nil {
		return entity.UserTransaction{}, fmt.Errorf("exec tx: %w", err)
	}

	// Сформировать JSON-событие аудита
	go func() {
		auditEvent := entity.NewTransactionAuditEventFromTransaction(resultTxn)
		payload, errMarshal := json.Marshal(auditEvent)
		if errMarshal != nil {
			// Можно залогировать ошибку сериализации, но не блокировать основную операцию
			return
		} else {
			if err := r.SaveAuditLog(resultTxn.ID, string(payload), auditEvent.Version, time.Now()); err != nil {
				//r.logger.Error("failed to save transaction audit log", "error", err.Error(), "txn_id", resultTxn.ID)
			}
		}
	}()

	return resultTxn, nil
}

// lockUserBalance блокирует строку пользователя в таблице users и возвращает его текущий баланс.
func (q *queries) lockUserBalance(ctx context.Context, userID int) (int, error) {
	query := `
		SELECT balance
		FROM user_accounts
		WHERE user_id = $1
		FOR UPDATE
	`
	var balance int
	err := q.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err == sql.ErrNoRows {
		_, err2 := q.db.ExecContext(ctx, "INSERT INTO user_accounts (user_id, balance) VALUES ($1, $2)", userID, 0)
		if err2 != nil {
			return 0, fmt.Errorf("failed to create user account: %w", err2)
		}
	} else if err != nil {
		return 0, fmt.Errorf("failed to lock user balance: %w", err)
	}
	return balance, nil
}

// insertTransaction вставляет новую транзакцию и возвращает созданную запись.
func (q *queries) insertTransaction(ctx context.Context, txn entity.UserTransaction) (entity.UserTransaction, error) {
	query := `
		INSERT INTO user_transactions (
			user_id, type, amount, status, provider, external_id,
			service_id, bot_id, description, promo_code_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,
			CASE WHEN $10 <> '' THEN (SELECT code FROM promo_codes WHERE code = $10) ELSE NULL END
		)
		RETURNING id, created_at, updated_at
	`
	promoCode := ""
	if txn.PromoCode != nil {
		promoCode = txn.PromoCode.Code
	}

	serviceID := sql.NullInt64{}
	if txn.ServiceID != 0 {
		serviceID = sql.NullInt64{Int64: int64(txn.ServiceID), Valid: true}
	}

	err := q.db.QueryRowContext(ctx, query,
		txn.UserID,
		txn.Type,
		txn.Amount,
		txn.Status,
		txn.Provider,
		txn.ExternalID,
		serviceID,
		txn.BotID,
		txn.Description,
		promoCode,
	).Scan(&txn.ID, &txn.CreatedAt, &txn.UpdatedAt)
	if err != nil {
		return entity.UserTransaction{}, fmt.Errorf("failed to insert transaction: %w", err)
	}
	return txn, nil
}

func (q *queries) lockTransaction(ctx context.Context, txnID string) (userID int, txnType string, amount int, err error) {
	query := "SELECT user_id, type, amount FROM user_transactions WHERE id = $1 FOR UPDATE"
	err = q.db.QueryRowContext(ctx, query, txnID).Scan(&userID, &txnType, &amount)
	if err != nil {
		return 0, "", 0, fmt.Errorf("failed to lock transaction: %w", err)
	}
	return userID, txnType, amount, nil
}

func (q *queries) updateTransactionStatus(ctx context.Context, txnID string, status string) error {
	_, err := q.db.ExecContext(ctx, "UPDATE user_transactions SET status = $1, updated_at = NOW() WHERE id = $2", status, txnID)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}
	return nil
}

func (q *queries) updateUserBalance(ctx context.Context, userID int, balanceChange int) error {
	var currentBalance int
	err := q.db.QueryRowContext(ctx, "SELECT balance FROM user_accounts WHERE user_id = $1 FOR UPDATE", userID).Scan(&currentBalance)
	if err != nil {
		if err == sql.ErrNoRows {
			currentBalance = 0
			// Создаем новую запись с балансом 0, если пользователь не найден
			_, err2 := q.db.ExecContext(ctx, "INSERT INTO user_accounts (user_id, balance) VALUES ($1, $2)", userID, currentBalance)
			if err2 != nil {
				return fmt.Errorf("failed to create user account: %w", err2)
			}
		} else {
			return fmt.Errorf("failed to lock user account: %w", err)
		}
	}
	newBalance := currentBalance + balanceChange
	if newBalance < 0 {
		return fmt.Errorf("insufficient funds: balance would become negative")
	}
	_, err = q.db.ExecContext(ctx, "UPDATE user_accounts SET balance = $1 WHERE user_id = $2", newBalance, userID)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}
	return nil
}

func (r *Repository) UpdateTransactionExternalID(ctx context.Context, txnID string, externalID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE user_transactions SET external_id = $1 WHERE id = $2", externalID, txnID)
	if err != nil {
		return fmt.Errorf("failed to update transaction external ID: %w", err)
	}
	return nil
}

func (r *Repository) SaveAuditLog(txnID string, payload string, version int, eventtime time.Time) error {
	return r.audit.SaveAudit(txnID, payload, version, eventtime)
}

func (q *queries) saveAuditLog(txnID string, payload string, version int, eventtime time.Time) error {
	return q.audit.SaveAudit(txnID, payload, version, eventtime)
}
