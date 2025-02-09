package elysium

import (
	"context"
	"database/sql"
	"elysium/internal/entity"
	"fmt"
	"strings"
)

func (r *Repository) CreatePendingTransactions(ctx context.Context, userID int, transactions []entity.PendingTransactionInput) ([]entity.UserTransaction, error) {
	if len(transactions) == 0 {
		return []entity.UserTransaction{}, nil
	}

	valueStrings := make([]string, 0, len(transactions))
	valueArgs := make([]interface{}, 0, len(transactions)*3+1)
	valueArgs = append(valueArgs, userID)

	for i, txn := range transactions {
		valueStrings = append(valueStrings, fmt.Sprintf("($1::integer, $%d::text, $%d::integer, $%d::text)", i*3+2, i*3+3, i*3+4))
		valueArgs = append(valueArgs, txn.Type)
		valueArgs = append(valueArgs, txn.Amount)
		valueArgs = append(valueArgs, txn.Provider)
	}

	query := fmt.Sprintf(`
		WITH locked_user AS (
			SELECT id, balance
			FROM users
			WHERE id = $1
			FOR UPDATE
		),
		ins AS (
			INSERT INTO user_transactions (
				user_id, type, amount, provider, status, balance_after
			)
			SELECT
				v.user_id,
				v.type,
				v.amount,
				v.provider,
				'pending',
				CASE v.type::text
					WHEN 'deposit' THEN lu.balance + v.amount::integer
					WHEN 'withdrawal' THEN lu.balance - v.amount::integer
					WHEN 'refund' THEN lu.balance + v.amount::integer
					ELSE lu.balance
				END as balance_after
			FROM (VALUES %s) as v(user_id, type, amount, provider)
			CROSS JOIN locked_user lu
			RETURNING id, amount, created_at, updated_at, balance_after
		)
		SELECT id, amount, created_at, updated_at, balance_after FROM ins
	`, strings.Join(valueStrings, ","))

	rows, err := r.db.QueryContext(ctx, query, valueArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to create pending transactions: %w", err)
	}
	defer rows.Close()

	var result []entity.UserTransaction
	for rows.Next() {
		var txn entity.UserTransaction
		var amount int
		err := rows.Scan(&txn.ID, &amount, &txn.CreatedAt, &txn.UpdatedAt, &txn.BalanceAfter)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		txn.UserID = userID
		txn.Status = entity.TransactionStatusPending
		txn.Amount = amount
		result = append(result, txn)
	}

	return result, nil
}

func (r *Repository) CreateTransaction(ctx context.Context, txn entity.UserTransaction) (entity.UserTransaction, error) {
	query := `                                                                                                                                                                                                    
         WITH locked_user AS (                                                                                                                                                                                     
              SELECT id, balance                                                                                                                                                                                    
              FROM users                                                                                                                                                                                            
              WHERE id = $1                                                                                                                                                                                         
              FOR UPDATE                                                                                                                                                                                            
          ),                                                                                                                                                                                                        
          new_balance AS (                                                                                                                                                                                          
              SELECT
                  CASE
                      WHEN $4 = 'pending' THEN balance
                      WHEN $2 = 'deposit' THEN balance + $3
                      WHEN $2 = 'withdrawal' THEN balance - $3
                      WHEN $2 = 'refund' THEN balance + $3
                  END AS balance
              FROM locked_user
          ),                                                                                                                                                                                                        
          insert_txn AS (                                                                                                                                                                                           
              INSERT INTO user_transactions (                                                                                                                                                                       
                  user_id, type, amount, status, provider, external_id,                                                                                                                                             
                  service_id, bot_id, balance_after, description                                                                                                                                                    
              )                                                                                                                                                                                                     
              SELECT                                                                                                                                                                                                
                  $1, $2, $3, $4, $5, $6, $7, $8, nb.balance, $9                                                                                                                                                    
              FROM new_balance nb                                                                                                                                                                                   
              RETURNING                                                                                                                                                                                             
                  id, created_at, updated_at, balance_after                                                                                                                                                         
          )                                                                                                                                                                                                         
          UPDATE users u                                                                                                                                                                                            
          SET balance = ib.balance_after                                                                                                                                                                            
          FROM insert_txn ib                                                                                                                                                                                        
          WHERE u.id = $1                                                                                                                                                                                           
          RETURNING ib.id, ib.created_at, ib.updated_at, ib.balance_after                                                                                                                                           
      `

	err := r.db.QueryRowContext(ctx, query,
		txn.UserID,
		txn.Type,
		txn.Amount,
		txn.Status,
		txn.Provider,
		txn.ExternalID,
		txn.ServiceID,
		txn.BotID,
		txn.Description,
	).Scan(
		&txn.ID,
		&txn.CreatedAt,
		&txn.UpdatedAt,
		&txn.BalanceAfter,
	)

	if err != nil {
		return entity.UserTransaction{}, fmt.Errorf("transaction failed: %w", err)
	}

	return txn, nil
}

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
			t.balance_after,
			t.description,
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
		var provider, externalID sql.NullString

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
			&txn.BalanceAfter,
			&txn.Description,
			&txn.CreatedAt,
			&txn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		if serviceID.Valid {
			txn.ServiceID = new(int)
			*txn.ServiceID = int(serviceID.Int64)
		}
		if botID.Valid {
			txn.BotID = new(int64)
			*txn.BotID = botID.Int64
		}
		if provider.Valid {
			txn.Provider = provider.String
		}
		if externalID.Valid {
			txn.ExternalID = externalID.String
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
			t.balance_after,
			t.description,
			t.created_at,
			t.updated_at
		FROM user_transactions t
		LEFT JOIN services s ON t.service_id = s.id
		WHERE t.id = $1
	`

	var txn entity.UserTransaction
	var serviceID sql.NullInt64
	var botID sql.NullInt64
	var provider, externalID, description sql.NullString

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
		&txn.BalanceAfter,
		&description,
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
		txn.ServiceID = new(int)
		*txn.ServiceID = int(serviceID.Int64)
	}
	if botID.Valid {
		txn.BotID = new(int64)
		*txn.BotID = botID.Int64
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

	return txn, nil
}

func (r *Repository) UpdateTransactionStatus(ctx context.Context, txnID string, status string) error {
	query := `
		WITH locked_txn AS (
			SELECT user_id, type, amount
			FROM user_transactions
			WHERE id = $1
			FOR UPDATE
		),
		locked_user AS (
			SELECT id, balance
			FROM users
			WHERE id = (SELECT user_id FROM locked_txn)
			FOR UPDATE
		),
		new_balance AS (
			SELECT
				CASE
					WHEN lt.type = 'deposit' THEN lu.balance + lt.amount
					WHEN lt.type = 'withdrawal' THEN lu.balance - lt.amount
					WHEN lt.type = 'refund' THEN lu.balance + lt.amount
					ELSE lu.balance
				END AS balance
			FROM locked_txn lt
			CROSS JOIN locked_user lu
		)
		UPDATE user_transactions ut
		SET 
			status = $2,
			balance_after = nb.balance,
			updated_at = NOW()
		FROM new_balance nb
		WHERE ut.id = $1
		RETURNING ut.user_id
	`

	var userID int
	err := r.db.QueryRowContext(ctx, query, txnID, status).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	return nil
}
