# Архитектура и назначение проекта

Проект представляет собой бэкенд-сервис для управления Telegram-ботами с интеграцией:

- Веб-интерфейс (HTTP-контроллеры)
- Работа с аудиотреками (SoundCloud, загрузка песен)
- Система кеширования (Redis)
- СУБД (Postgres, ClickHouse)
- Планировщик задач (gocron)
- Система прав пользователей и транзакций
- Генерация эмодзи-паков

## Основные модули и файловая структура

### 1. Конфигурация (/internal/application/config/)

- Модули: ArimaDJ, Clickhouse, Demethra, Telegram и др.
- Структура:

```go
// config.go
type Module struct {
    configs []any  // Все конфигурационные структуры
}

// Инициализация конфигурации из JSON
func (m Module) Init(st interface{ Config() ([]byte, error) }) error {
    // Парсинг конфигурации для каждого модуля
}
```

### 2. Контроллеры (/internal/controller/)

#### HTTP-роутинг (/http/):
```go
type Module struct {
    server       *http.Server
    ctx          context.Context
    stop         context.CancelFunc
    logger       *slog.Logger
    groups       []group
    certFilePath string
    keyFilePath  string
}
```

#### Планировщик (/scheduler/):
```go
type Module struct {
    ctx             context.Context
    stop            context.CancelFunc
    logger          *slog.Logger
    scheduler       *gocron.Scheduler
    reconfiguration chan struct{}
    jobs            []job
}
```

#### Telegram-боты (/telegram/):
```go
type Module struct {
    ctx         context.Context
    stop        context.CancelFunc
    logger      *slog.Logger
    cfg         config
    bot         *telego.Bot
    updates     <-chan telego.Update
    botHandler  *telegohandler.BotHandler
    middleware  []Middleware
    commands    []Command
    groupHandle []GroupHandle
    handles     []Handle
}
```

### 3. Сущности (/internal/entity/)

```go
// User - основная сущность пользователя
type User struct {
    ID               int         `db:"id" json:"id"`
    TelegramID       int64      `db:"telegram_id" json:"telegram_id"`
    TelegramUsername string     `db:"telegram_username" json:"username"`
    Firstname        string     `db:"firstname" json:"firstname"`
    Permissions      Permissions `db:"permissions" json:"permissions"`
    BotsActivated    []*Bot     `db:"bots_activated" json:"bots_activated"`
    Balance          int        `db:"balance" json:"balance"`
    DateCreate       time.Time  `db:"date_create" json:"date_create"`
}
```

### 4. Репозитории (/internal/repository/)

#### Postgres (/postgres/):
```go
type Module struct {
    cfg    config
    db     *sqlx.DB
    tables []table
}
```

#### Redis (/redis/):
```go
type Module struct {
    cfg     config
    client  *redis.Client
    logger  *slog.Logger
}
```

### 5. Юзкейсы (/internal/usecase/)

#### Управление пользователями (/users/):
```go
type Module struct {
    logger           *slog.Logger
    ctx             context.Context
    cache           cacheUC
    repo            repository
    onlineUsersCount atomic.Int64
    mu              sync.RWMutex
    streamsOnlineCount map[string]int64
}
```

#### Аутентификация (/auth/):
```go
type Module struct {
    ctx       context.Context
    cfg       config
    logger    *slog.Logger
    jwtSecret []byte
    mu        sync.RWMutex
    tokensMap sync.Map
    cache     externalCache
    repo      repository
    users     userCreator
}
```

## Ключевые особенности архитектуры

1. **Слоистая структура:**
   - Контроллеры → Юзкейсы → Репозитории → Сущности
   - Четкое разделение ответственности между слоями

2. **Dependency Injection:**
   - Зависимости передаются явно через конструкторы
   - Использование интерфейсов для слабого связывания компонентов

3. **Гибкая конфигурация:**
   - Поддержка разных окружений (local/test/prod)
   - JSON-based конфигурация для всех модулей

4. **Интеграции:**
   - Telegram Bot API (telego)
   - Валидация конфигов (ozzo-validation)
   - Планировщик задач (gocron)
   - TLS поддержка для HTTP сервера

5. **Отказоустойчивость:**
   - Контекст-based отмена операций
   - Graceful shutdown для всех сервисов
   - Middleware для восстановления после паники
