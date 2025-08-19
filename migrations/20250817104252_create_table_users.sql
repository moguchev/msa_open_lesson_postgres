-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id         UUID NOT NULL DEFAULT gen_random_uuid(),         -- Уникальный идентификатор ползователя
    email      TEXT NOT NULL CHECK (email = lower(email)),      -- email ползователя
    username   TEXT NOT NULL,                                   -- username ползователя
    full_name  TEXT,                                            -- Полное имя ползователя
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), -- Когда создан
    last_login TIMESTAMP WITH TIME ZONE,                        -- Последний вход
    is_active  BOOLEAN NOT NULL DEFAULT true,                   -- Активный пользователь
    PRIMARY KEY(id),
    UNIQUE(email)
);

COMMENT ON TABLE users IS 'Таблица пользователей';

COMMENT ON COLUMN users.id IS 'Уникальный идентификатор ползователя';
COMMENT ON COLUMN users.email IS 'email ползователя';
COMMENT ON COLUMN users.username IS 'username ползователя';
COMMENT ON COLUMN users.full_name IS 'Полное имя ползователя';
COMMENT ON COLUMN users.created_at IS 'Когда создан';
COMMENT ON COLUMN users.last_login IS 'Последний вход';
COMMENT ON COLUMN users.is_active IS 'Активный пользователь';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
