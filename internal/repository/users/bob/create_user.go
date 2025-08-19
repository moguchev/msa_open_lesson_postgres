package users

import (
	"context"
	"errors"

	"github.com/moguchev/msa_open_lesson_postgres/internal/gen/bob/dberrors"
	"github.com/moguchev/msa_open_lesson_postgres/internal/gen/bob/schema"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
)

// ORM

func (r *Repository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	createdUser, err := schema.Users.Insert(userToSchema(user)).One(ctx, r.bobDB)
	if err != nil {
		if errors.Is(err, dberrors.ErrUniqueConstraint) {
			return nil, errors.New("already exists")
		}
		return nil, err
	}

	return userFromSchema(createdUser), nil
}
