-- Включаем UUID расширение для генерации уникальных идентификаторов
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Создаем таблицу пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    surname TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL,
    salt TEXT NOT NULL,
    avatar TEXT,
    phone TEXT,
    bio TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем таблицу мероприятий
CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    location TEXT NOT NULL,
    latitude FLOAT,
    longitude FLOAT, 
    category TEXT,
    date_start TIMESTAMP NOT NULL,
    date_end TIMESTAMP NOT NULL,
    organizer_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    image_url TEXT,
    price DECIMAL(10, 2),
    max_participants INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создаем таблицу тегов мероприятий
CREATE TABLE IF NOT EXISTS event_tags (
    id SERIAL PRIMARY KEY,
    event_id INTEGER REFERENCES events(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    UNIQUE (event_id, tag)
);

-- Таблица связи пользователей и событий (участники мероприятий)
CREATE TABLE IF NOT EXISTS event_participants (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT DEFAULT 'registered', -- registered, attended, cancelled
    registration_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (event_id, user_id)
);

-- Таблица для комментариев к мероприятиям
CREATE TABLE IF NOT EXISTS event_comments (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id INTEGER REFERENCES event_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица для оценок мероприятий
CREATE TABLE IF NOT EXISTS event_ratings (
    id SERIAL PRIMARY KEY,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
    review TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (event_id, user_id)
);

-- Таблица для избранных мероприятий пользователя
CREATE TABLE IF NOT EXISTS user_favorite_events (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    event_id INTEGER NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, event_id)
);

-- Таблица для уведомлений пользователей
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    type TEXT NOT NULL, -- event_reminder, system, etc.
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Создание индексов для оптимизации запросов

-- Индекс для быстрого поиска по email
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Индекс для ускорения поиска по роли 
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Индекс для быстрого поиска мероприятий по категории
CREATE INDEX IF NOT EXISTS idx_events_category ON events(category);

-- Индексы для работы с геопозицией
CREATE INDEX IF NOT EXISTS idx_events_location_geo ON events(latitude, longitude);

-- Индекс для поиска по дате мероприятий
CREATE INDEX IF NOT EXISTS idx_events_date ON events(date_start, date_end);

-- Индекс для быстрого доступа к мероприятиям по организатору
CREATE INDEX IF NOT EXISTS idx_events_organizer ON events(organizer_id);

-- Индекс для поиска по тегам
CREATE INDEX IF NOT EXISTS idx_event_tags_tag ON event_tags(tag);

-- Индекс для быстрого доступа к комментариям мероприятия
CREATE INDEX IF NOT EXISTS idx_event_comments_event ON event_comments(event_id);

-- Индекс для быстрого доступа к участникам мероприятия
CREATE INDEX IF NOT EXISTS idx_event_participants_event ON event_participants(event_id);
CREATE INDEX IF NOT EXISTS idx_event_participants_user ON event_participants(user_id);

-- Индекс для быстрого доступа к избранным мероприятиям пользователя
CREATE INDEX IF NOT EXISTS idx_user_favorite_events_user ON user_favorite_events(user_id);

-- Индекс для быстрого доступа к непрочитанным уведомлениям
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id) WHERE read = FALSE;

-- Функция автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_users_modtime
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE PROCEDURE update_modified_column();
    
CREATE TRIGGER update_events_modtime
    BEFORE UPDATE ON events
    FOR EACH ROW
    EXECUTE PROCEDURE update_modified_column();
    
CREATE TRIGGER update_event_comments_modtime
    BEFORE UPDATE ON event_comments
    FOR EACH ROW
    EXECUTE PROCEDURE update_modified_column();

-- Функция для очистки старых уведомлений (будет запускаться по расписанию)
CREATE OR REPLACE FUNCTION cleanup_old_notifications()
RETURNS void AS $$
BEGIN
    -- Удаляет прочитанные уведомления старше 90 дней
    DELETE FROM notifications
    WHERE read = TRUE
    AND created_at < (CURRENT_TIMESTAMP - INTERVAL '90 days');
END;
$$ LANGUAGE plpgsql; 