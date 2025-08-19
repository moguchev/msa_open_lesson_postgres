package users

import (
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stephenafamo/bob"
)

type Repository struct {
	sqlDB *sql.DB
	bobDB bob.DB
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return &Repository{
		sqlDB: sqlDB,
		bobDB: bob.NewDB(sqlDB),
	}
}
