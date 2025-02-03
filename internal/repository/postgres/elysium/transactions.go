package elysium

import (
	"context"
	"database/sql"
	"elysium/internal/entity"
	"fmt"
)

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
                 CASE $2                                                                                                                                                                                           
                     WHEN 'deposit' THEN balance + $3                                                                                                                                                              
                     WHEN 'withdrawal' THEN balance - $3                                                                                                                                                           
                     WHEN 'refund' THEN balance + $3                                                                                                                                                               
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
