# msa_open_lesson_postgres
Открытый урок MSA Postgres

## PostgreSQL

1. Поднимаем контейнер с PostgreSQL (см. [docker-compose.yaml](./docker-compose.yaml#9))
```sh
docker compose up -d 
```

## Миграции

1. Устанавливаем [goose](https://github.com/pressly/goose):
```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -version
```

2. Создаем миграцию создания таблицы пользователей:
```sh
goose create create_table_users sql -dir=./migrations
```
```sql
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
```

3. 
```sh
goose create create_indexes_table_users sql -dir=./migrations
```
```sql
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
```

4. Проверяем статус миграций
```sh
goose postgres "user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} host=127.0.0.1 port=5432 sslmode=disable" status -dir ./migrations
```
```log
2025/08/17 14:26:22     Applied At                  Migration
2025/08/17 14:26:22     =======================================
2025/08/17 14:26:22     Pending                  -- 20250817104252_create_table_users.sql
2025/08/17 14:26:22     Pending                  -- 20250817104936_create_indexes_table_users.sql
```

5. Накатываем миграции
```sh
goose postgres "user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} host=127.0.0.1 port=5432 sslmode=disable" up -dir ./migrations
```
```log
2025/08/17 14:29:14 OK   20250817104252_create_table_users.sql (10.5ms)
2025/08/17 14:29:14 OK   20250817104936_create_indexes_table_users.sql (11.68ms)
2025/08/17 14:29:14 goose: successfully migrated database to version: 20250817104936
```

6. Повторно проверяем статус миграций
```sh
goose postgres "user=${POSTGRES_USER} password=${POSTGRES_PASSWORD} dbname=${POSTGRES_DB} host=127.0.0.1 port=5432 sslmode=disable" status -dir ./migrations
```
```log
2025/08/17 14:32:51     Applied At                  Migration
2025/08/17 14:32:51     =======================================
2025/08/17 14:32:51     Sun Aug 17 11:29:14 2025 -- 20250817104252_create_table_users.sql
2025/08/17 14:32:51     Sun Aug 17 11:29:14 2025 -- 20250817104936_create_indexes_table_users.sql
```


## Bob

[Bob - Go SQL Access Toolkit](https://bob.stephenafamo.com/)

1. Устанавливаем `bobgen-psql`
```sh
go get -tool github.com/stephenafamo/bob/gen/bobgen-psql@latest 
```

### Кодогенерация ORM

1. Создадим файл [bobgen.yaml](./config/bobgen.yaml):
```yaml
# Подробнее: https://bob.stephenafamo.com/docs/code-generation/psql#driver-configuration
psql:
  dsn: "postgres://example_user:example_password@127.0.0.1:5432/example?sslmode=disable"
  driver: "github.com/jackc/pgx/v5/stdlib"
  schemas:
    - "public"
  uuid_pkg: "google"
  except:
    goose_db_version: []

# Подробнее: https://bob.stephenafamo.com/docs/code-generation/configuration#plugins-configuration
plugins:
  dbinfo: # Генерирует код для получения информации о каждой базе данных: схемах, таблицах, столбцах, индексах, первичных ключах, внешних ключах, уникальных ограничениях и проверочных ограничениях.
    disabled: false
    pkgname: "dbinfo"
    destination: "internal/gen/bob/dbinfo"
  enums: # Генерирует код для перечислений в отдельном пакете, если таковые имеются.
    disabled: false
    pkgname: "enums"
    destination: "internal/gen/bob/enums"
  models: # Генерирует код для моделей. Зависит от перечислений.
    disabled: false
    pkgname: "schema"
    destination: "internal/gen/bob/schema"
  factory: # Генерирует код для фабрик. Зависит от моделей.
    disabled: false
    pkgname: "factory"
    destination: "internal/gen/bob/factory"
  dberrors: # Генерирует код для ошибок уникальных ограничений. Зависит от моделей.
    disabled: false
    pkgname: "dberrors"
    destination: "internal/gen/bob/dberrors"
  where: # Добавляет шаблоны в пакет моделей для генерации кода для предложений WHERE, например
    disabled: false
  loaders: # Добавляет шаблоны в пакет моделей для генерации кода для загрузчиков, например models.SelectThenLoad.Table.Rel()
    disabled: false
  joins:
    disabled: false
```

2. Запустим утиллиту **bobgen-psql**
```sh
go tool github.com/stephenafamo/bob/gen/bobgen-psql -c ./config/bobgen.yaml
```

3. в пакете `internal/gen/bob` появятся сгенерированные файлы


### Кодогенерация Запросов


1. Создадим файл [queries.sql](./internal/repository/users/bob/sql/queries.sql):
```sql
-- FindUserByEmail
SELECT * FROM users WHERE email = $1;
```

2. В [bobgen.yaml](./config/bobgen.yaml) укажем путь до файла с запросами:
```yaml
psql:
  # ...
  queries:
    - ./internal/repository/users/bob/sql
```

3. Запустим утиллиту **bobgen-psql**
```sh
go tool github.com/stephenafamo/bob/gen/bobgen-psql -c ./config/bobgen.yaml
```

4. В директории `/internal/repository/users/bob/sql` появятся сгенерированные файлы