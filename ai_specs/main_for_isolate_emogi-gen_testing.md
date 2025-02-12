# High-Level Objective
- Добавление промокодов.

# Implementation Notes


# Low-level tasks
> ordered from start to finish


1. UPDATE 20250209150225_transactions.sql: Новая таблица promo_codes:
   • code – уникальный код промокода (тип VARCHAR или другой по договорённости);
   • type – тип промокода (например, 'referral', 'single', 'free_use' и т.п.);
   Рекомендуется добавить CHECK-констрейнт для допустимых значений.
   • user_id – идентификатор пользователя-владельца (если применяется реферальный код), внешний ключ к users(id);
   • bonus_redeemer – количество бонусных токенов, начисляемых тому, кто применил промокод;
   • bonus_referrer – количество бонусов для владельца промокода (если применимо);
   • usage_limit – максимально допустимое число использований кодом;
   • usage_count – счетчик использования, чтобы отслеживать оставшееся количество применений;
   • valid_from и valid_to – период действия промокода;
   • Стандартные поля created_at и updated_at.

2. Расширение таблицы транзакций user_transactions:
• Расширение поля type:
- 'promo_redeem' – для начисления бонуса при применении промокода;
- 'referral_bonus' – для начисления бонуса владельцу реферального кода.
• Новый столбец promo_code_id который будет ссылаться на таблицу promo_codes.

3. Создание таблицы для отслеживания использований промокодов promo_code_usages



у меня есть метод для выполнения транзакций `execTX(ctx context.Context, fn func(*queries) error) error`. Вот пример его использования:
```
err := r.execTX(ctx, func(q *queries) error {
 var err error
 user, err = q.getUserByTelegramUserID(ctx, telegramID)
 if err != nil {
 return fmt.Errorf("failed to get user: %w", err)
 }
bots, err := q.getUserActiveBots(ctx, user.ID)
if err != nil {
return fmt.Errorf("get user active bots: %w", err)
}

 user.BotsActivated = bots

 return nil

 })
 if err != nil {
 return entity.User{}, fmt.Errorf("exec tx: %w", err)

}
```


1. Новый агрегирующий метод.
Создайте новый метод в вашем репозитории ProcessTransaction, который внутри себя вызывает обновление статуса транзакции и обновление баланса. Оба действия должны выполняться
через один вызов execTX.                                                                                                                                                                                        
2. Использовать один транзакционный блок.
Вместо отдельных вызовов UpdateTransactionStatus и UpdateUserBalance, объедините их следующим образом:
```
    func (r *Repository) ProcessTransaction(ctx context.Context, txnID string, userID int, balanceChange int, newStatus string) error {                                                                           
        return r.execTX(ctx, func(q *queries) error {                                                                                                                                                              
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
    }                                                                                                                                                                                                              
```
   Здесь параметр balanceChange можно вычислить заранее, исходя из типа транзакции и суммы.  

3. Метод обновления баланса в транзакционном контексте.
   Если у вас уже есть метод updateUserBalance в типе queries, убедитесь, что он использует запросы с FOR UPDATE для получения актуального состояния баланса и выполняет обновление внутри той же транзакции.      
 
