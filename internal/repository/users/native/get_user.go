package users

import (
	"context"
	"errors"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
)

func (r *Repository) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := r.sb.
		Select(usersTableColumns...).
		From(usersTable).
		Where(squirrel.Eq{usersTableColumnID: id})

	var row UserRow
	if err := r.pool.Getx(ctx, &row, query); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("not found")
		}
		return nil, err
	}

	return ToModel(&row), nil
}
