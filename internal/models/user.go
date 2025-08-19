package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Username  string
	FullName  string
	CreatedAt time.Time
	LastLogin time.Time
	IsActive  bool
}

type UserFilter struct {
	IDs         []uuid.UUID
	Email       *string // ищем как точное/ILIKE
	Username    *string // ILIKE %...%
	FullText    *string // tsquery
	IsActive    *bool
	CreatedFrom *time.Time
	CreatedTo   *time.Time
}
