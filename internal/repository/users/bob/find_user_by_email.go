package users

import (
	"context"
	"strings"

	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
	sql "github.com/moguchev/msa_open_lesson_postgres/internal/repository/users/bob/sql"
)

// Code generate

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := sql.FindUserByEmail(strings.ToLower(email)).One(ctx, r.bobDB)
	if err != nil {
		return nil, err
	}
	return userFromFindUserByEmailRow(&user), nil
}
