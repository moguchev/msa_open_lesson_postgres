package users

import (
	"context"
	"strings"

	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
)

func (r *Repository) CreateUser(ctx context.Context, in *models.User) (*models.User, error) {
	row := FromModel(in)

	query := r.sb.
		Insert(usersTable).
		Columns(usersTableColumns...).
		Values(row.Values()...).
		Suffix("RETURNING " + strings.Join(usersTableColumns, ","))

	var outRow UserRow
	if err := r.pool.Getx(ctx, &outRow, query); err != nil {
		return nil, err
	}

	return ToModel(&outRow), nil
}
