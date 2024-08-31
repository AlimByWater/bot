-- Создание таблицы для хранения макетов пользователей
CREATE TABLE IF NOT EXISTS elysium.user_layouts (
                                                    name VARCHAR(255) NOT NULL,
                                                    layout_id VARCHAR(255) PRIMARY KEY,
                                                    creator INT NOT NULL,
                                                    date_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для хранения фонов макетов
CREATE TABLE IF NOT EXISTS elysium.backgrounds (
                                                   layout_id VARCHAR(255) PRIMARY KEY REFERENCES elysium.user_layouts(layout_id) ON DELETE CASCADE,
                                                   type VARCHAR(50) NOT NULL,
                                                   value VARCHAR(255) NOT NULL
);

-- Создание таблицы для хранения элементов макетов
CREATE TABLE IF NOT EXISTS elysium.layout_elements (
                                                       element_id VARCHAR(255) PRIMARY KEY,
                                                       layout_id VARCHAR(255) NOT NULL REFERENCES elysium.user_layouts(layout_id) ON DELETE CASCADE,
                                                       type VARCHAR(50) NOT NULL,
                                                       public BOOLEAN NOT NULL,
                                                       removable BOOLEAN NOT NULL
);

-- Создание таблицы для хранения позиций элементов макетов
CREATE TABLE IF NOT EXISTS elysium.positions (
                                                 element_id VARCHAR(255) PRIMARY KEY REFERENCES elysium.layout_elements(element_id) ON DELETE CASCADE,
                                                 row INT NOT NULL,
                                                 col INT NOT NULL,
                                                 height INT NOT NULL,
                                                 width INT NOT NULL
);

-- Создание таблицы для хранения свойств элементов макетов
CREATE TABLE IF NOT EXISTS elysium.properties (
                                                  element_id VARCHAR(255) PRIMARY KEY REFERENCES elysium.layout_elements(element_id) ON DELETE CASCADE,
                                                  icon VARCHAR(255),
                                                  title VARCHAR(255),
                                                  navigation_url VARCHAR(255),
                                                  current_value INT,
                                                  min_value INT,
                                                  max_value INT,
                                                  value INT
);

-- Создание таблицы для хранения редакторов макетов
CREATE TABLE IF NOT EXISTS elysium.layout_editors (
                                                      layout_id VARCHAR(255) NOT NULL REFERENCES elysium.user_layouts(layout_id) ON DELETE CASCADE,
                                                      editor_id INT NOT NULL REFERENCES elysium.users(id) ON DELETE CASCADE
);

-- Создание индексов для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_user_layouts_creator_id ON elysium.user_layouts(creator);
CREATE INDEX IF NOT EXISTS idx_layout_elements_layout_id ON elysium.layout_elements(layout_id);
CREATE INDEX IF NOT EXISTS idx_layout_editors_layout_id ON elysium.layout_editors(layout_id);