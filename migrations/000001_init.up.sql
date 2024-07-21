-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS elysium.users (
                       id SERIAL PRIMARY KEY,
                       telegram_id bigint UNIQUE ,
                       telegram_username VARCHAR(255),
                       firstname VARCHAR(255),
                       date_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы песен
CREATE TABLE IF NOT EXISTS elysium.songs (
                       id SERIAL PRIMARY KEY,
                       url VARCHAR(255) UNIQUE NOT NULL CHECK (url <> ''),
                       artist_name VARCHAR(255) NOT NULL CHECK (artist_name <> ''),
                       title VARCHAR(255),
                       cover_link VARCHAR(255),
                       cover VARCHAR(255),
                       cover_telegram_file_id VARCHAR(255),
                       song_telegram_message_chat_id bigint,
                       song_telegram_message_id bigint,
                       tags TEXT[],
                       date_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для записи загрузок песен пользователями
CREATE TABLE IF NOT EXISTS elysium.user_song_downloads (
                                     id SERIAL PRIMARY KEY,
                                     user_id INT NOT NULL REFERENCES elysium.users(id) ON DELETE CASCADE,
                                     song_id INT NOT NULL REFERENCES elysium.songs(id) ON DELETE CASCADE,
                                     download_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для записи воспроизведения песен
CREATE TABLE IF NOT EXISTS elysium.song_plays (
                            id SERIAL PRIMARY KEY,
                            song_id INT NOT NULL REFERENCES elysium.songs(id) ON DELETE CASCADE,
                            play_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Создание таблицы для веб-приложения событий
CREATE TABLE IF NOT EXISTS elysium.web_app_events (
                          id SERIAL PRIMARY KEY,
                          event_type VARCHAR(50) NOT NULL,
                          user_id INT REFERENCES elysium.users(id),
                          telegram_user_id BIGINT NOT NULL REFERENCES elysium.users(telegram_id),
                          payload JSONB,
                          session_id VARCHAR(255) NOT NULL,
                          timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для быстрого поиска и уникальности
CREATE INDEX IF NOT EXISTS idx_users_username ON elysium.users(telegram_username);
CREATE INDEX IF NOT EXISTS idx_users_username ON elysium.users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_songs_title ON elysium.songs(title);
CREATE INDEX IF NOT EXISTS idx_songs_title ON elysium.songs(artist_name);
CREATE INDEX IF NOT EXISTS idx_songs_title ON elysium.songs(url);
CREATE INDEX IF NOT EXISTS idx_user_song_downloads_user_id ON elysium.user_song_downloads(user_id);
CREATE INDEX IF NOT EXISTS idx_user_song_downloads_song_id ON elysium.user_song_downloads(song_id);
CREATE INDEX IF NOT EXISTS idx_song_plays_song_id ON elysium.song_plays(song_id);
CREATE INDEX IF NOT EXISTS idx_web_app_events_user_id ON elysium.web_app_events(user_id);
CREATE INDEX IF NOT EXISTS idx_web_app_events_telegram_user_id ON elysium.web_app_events(telegram_user_id);
CREATE INDEX IF NOT EXISTS idx_web_app_events_session_id ON elysium.web_app_events(session_id);
CREATE INDEX IF NOT EXISTS idx_web_app_events_event_type ON elysium.web_app_events(event_type);



