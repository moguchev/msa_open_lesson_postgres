-- +goose Up
-- +goose NO TRANSACTION 

-- CREATE INDEX CONCURRENTLY cannot run inside a transaction block (SQLSTATE 25001)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_username 
ON users (username);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_full_name_gin 
ON users USING gin (to_tsvector('simple', full_name));

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_username, idx_users_full_name_gin;
-- +goose StatementEnd
