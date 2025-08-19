package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/moguchev/msa_open_lesson_postgres/internal/gen/bob/schema"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
)

// ORM

func (r *Repository) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user, err := schema.FindUser(ctx, r.bobDB, id)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			return nil, errors.New("not found")
		}
		return nil, err
	}

	return userFromSchema(user), nil
}
