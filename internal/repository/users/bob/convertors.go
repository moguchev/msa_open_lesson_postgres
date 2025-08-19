package users

import (
	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/moguchev/msa_open_lesson_postgres/internal/gen/bob/schema"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
	sql "github.com/moguchev/msa_open_lesson_postgres/internal/repository/users/bob/sql"
)

func userToSchema(user *models.User) *schema.UserSetter {
	return &schema.UserSetter{
		ID:        omit.From(user.ID),
		Email:     omit.From(user.Email),
		Username:  omit.From(user.Username),
		FullName:  omitnull.From(user.FullName),
		CreatedAt: omit.From(user.CreatedAt),
		LastLogin: omitnull.From(user.LastLogin),
		IsActive:  omit.From(user.IsActive),
	}
}

func userFromSchema(user *schema.User) *models.User {
	return &models.User{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		FullName:  user.FullName.GetOrZero(),
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin.GetOrZero(),
		IsActive:  user.IsActive,
	}
}

func userFromFindUserByEmailRow(user *sql.FindUserByEmailRow) *models.User {
	return &models.User{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		FullName:  user.FullName.GetOrZero(),
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLogin.GetOrZero(),
		IsActive:  user.IsActive,
	}
}
