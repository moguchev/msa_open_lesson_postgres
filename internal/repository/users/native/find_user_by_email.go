package users

import (
	"context"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
)

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, errors.New("empty email")
	}

	query := r.sb.
		Select(usersTableColumns...).
		From(usersTable).
		Where(sq.Eq{usersTableColumnEmail: email})

	var row UserRow
	if err := r.pool.Getx(ctx, &row, query); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}

	return ToModel(&row), nil
}
