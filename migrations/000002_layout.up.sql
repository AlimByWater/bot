-- Создание таблицы для корневых элементов
CREATE TABLE IF NOT EXISTS elysium.root_elements (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE CHECK (name <> ''),
    type TEXT NOT NULL CHECK (name <> ''),
    default_properties JSONB NOT NULL,
    description TEXT DEFAULT '',
    is_public BOOLEAN DEFAULT TRUE,
    is_paid BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для макетов пользователей
CREATE TABLE IF NOT EXISTS elysium.user_layouts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(250) UNIQUE NOT NULL CHECK (name <> ''),
    creator_id BIGINT NOT NULL REFERENCES elysium.users(id),
    stream_url VARCHAR(250) NOT NULL DEFAULT 'https://elysiumfm.ru/stream',
    background JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для редакторов макетов
CREATE TABLE IF NOT EXISTS elysium.layout_editors (
    id SERIAL PRIMARY KEY,
    layout_id INTEGER NOT NULL REFERENCES elysium.user_layouts(id) ON DELETE CASCADE,
    editor_id BIGINT NOT NULL REFERENCES elysium.users(id),
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(layout_id, editor_id)
);

-- Создание таблицы для элементов на макетах пользователей
CREATE TABLE IF NOT EXISTS elysium.layout_elements (
    id BIGSERIAL PRIMARY KEY,
    layout_id INTEGER NOT NULL REFERENCES elysium.user_layouts(id) ON DELETE CASCADE,
    root_element_id INTEGER NOT NULL REFERENCES elysium.root_elements(id),
    on_grid_id INTEGER NOT NULL DEFAULT 0,
    properties JSONB NOT NULL,
    position_x INTEGER NOT NULL,
    position_y INTEGER NOT NULL,
    position_z INTEGER NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    is_public BOOLEAN NOT NULL,
    is_removable BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для оптимизации запросов

CREATE INDEX IF NOT EXISTS idx_user_layouts_creator_id ON elysium.user_layouts(creator_id);
CREATE INDEX IF NOT EXISTS idx_user_layouts_name ON elysium.user_layouts(name);
CREATE INDEX IF NOT EXISTS idx_layout_editors_layout_id ON elysium.layout_editors(layout_id);
CREATE INDEX IF NOT EXISTS idx_layout_editors_editor_id ON elysium.layout_editors(editor_id);
CREATE INDEX IF NOT EXISTS idx_layout_elements_layout_id ON elysium.layout_elements(layout_id);
CREATE INDEX IF NOT EXISTS idx_layout_elements_properties ON elysium.layout_elements USING GIN (properties);

INSERT INTO elysium.root_elements (name, type, default_properties)
VALUES
('truchet', 'clickable_navigable', '{"value": 123}'),
('volume', 'clickable_non_navigable', '{"value": 123}'),
('online_count', 'non_clickable_non_navigable', '{"value": 123}');


-- Создание таблицы для логирования изменений макета
CREATE TABLE IF NOT EXISTS elysium.layout_changes (
    id SERIAL PRIMARY KEY,
    layout_id INTEGER NOT NULL REFERENCES elysium.user_layouts(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES elysium.users(telegram_id),
    change_type TEXT NOT NULL,
    details JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_layout_changes_layout_id ON elysium.layout_changes(layout_id);
CREATE INDEX IF NOT EXISTS idx_layout_changes_user_id ON elysium.layout_changes(user_id);
CREATE INDEX IF NOT EXISTS idx_layout_changes_timestamp ON elysium.layout_changes(timestamp);

-- Комментарии к таблице и столбцам для улучшения документации
COMMENT ON TABLE elysium.layout_changes IS 'Таблица для хранения истории изменений макетов';
COMMENT ON COLUMN elysium.layout_changes.id IS 'Уникальный идентификатор записи изменения';
COMMENT ON COLUMN elysium.layout_changes.layout_id IS 'ID макета, к которому относится изменение';
COMMENT ON COLUMN elysium.layout_changes.user_id IS 'ID пользователя, который внес изменение';
COMMENT ON COLUMN elysium.layout_changes.change_type IS 'Тип изменения (например, "create", "update", "delete")';
COMMENT ON COLUMN elysium.layout_changes.details IS 'Детали изменения'
