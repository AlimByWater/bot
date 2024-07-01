-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
                       id SERIAL PRIMARY KEY,
                       telegram_id bigint,
                       telegram_username VARCHAR(255),
                       firstname VARCHAR(255),
                       datecreate TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы песен
CREATE TABLE IF NOT EXISTS songs (
                       id SERIAL PRIMARY KEY,
                       url VARCHAR(255) NOT NULL,
                       artist_name VARCHAR(255) NOT NULL,
                       title VARCHAR(255) NOT NULL,
                       cover_link VARCHAR(255),
                       cover VARCHAR(255),
                       release_date DATE,
                       download_count INT DEFAULT 0,
                       tags TEXT[],
                       telegram_message_link VARCHAR(255) NOT NULL
);

-- Создание таблицы для записи загрузок песен пользователями
CREATE TABLE IF NOT EXISTS user_song_downloads (
                                     id SERIAL PRIMARY KEY,
                                     user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                     song_id INT NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
                                     download_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание таблицы для записи воспроизведения песен
CREATE TABLE IF NOT EXISTS song_plays (
                            id SERIAL PRIMARY KEY,
                            song_id INT NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
                            play_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для быстрого поиска и уникальности
CREATE INDEX IF NOT EXISTS idx_users_username ON users(telegram_username);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_songs_title ON songs(title);
CREATE INDEX IF NOT EXISTS idx_songs_title ON songs(artist_name);
CREATE INDEX IF NOT EXISTS idx_songs_title ON songs(url);
CREATE INDEX IF NOT EXISTS idx_user_song_downloads_user_id ON user_song_downloads(user_id);
CREATE INDEX IF NOT EXISTS idx_user_song_downloads_song_id ON user_song_downloads(song_id);
CREATE INDEX IF NOT EXISTS idx_song_plays_song_id ON song_plays(song_id);