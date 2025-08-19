package users

import (
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
)

// UserRow — «плоская» проекция строки таблицы users
// Nullable поля представлены sql.Null*.
type UserRow struct {
	ID        uuid.UUID      `db:"id"`
	Email     string         `db:"email"`
	Username  string         `db:"username"`
	FullName  sql.NullString `db:"full_name"`
	CreatedAt time.Time      `db:"created_at"`
	LastLogin sql.NullTime   `db:"last_login"`
	IsActive  bool           `db:"is_active"`
}

func (row *UserRow) Values() []any {
	return []any{
		row.ID, row.Email, row.Username, row.FullName, row.CreatedAt, row.LastLogin, row.IsActive,
	}
}

// ToModel конвертирует UserRow (sql.Null*) в доменную модель models.User (*string/*time.Time).
func ToModel(r *UserRow) *models.User {
	if r == nil {
		return nil
	}
	return &models.User{
		ID:        r.ID,
		Email:     r.Email,
		Username:  r.Username,
		FullName:  r.FullName.String,
		CreatedAt: r.CreatedAt,
		LastLogin: r.LastLogin.Time,
		IsActive:  r.IsActive,
	}
}

// FromModel конвертирует доменную модель в UserRow (для INSERT/UPDATE).
func FromModel(m *models.User) UserRow {
	if m == nil {
		return UserRow{}
	}
	return UserRow{
		ID:        m.ID,
		Email:     strings.ToLower(strings.TrimSpace(m.Email)),
		Username:  m.Username,
		FullName:  sql.NullString{String: m.FullName, Valid: m.FullName != ""},
		LastLogin: sql.NullTime{Time: m.LastLogin, Valid: !m.LastLogin.IsZero()},
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
	}
}
