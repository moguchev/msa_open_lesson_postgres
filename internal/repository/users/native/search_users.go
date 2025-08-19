package users

import (
	"context"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models/pagination"
)

func (r *Repository) SearchUsers(ctx context.Context, f *models.UserFilter, opts ...pagination.Option) ([]*models.User, error) {
	if f == nil {
		f = &models.UserFilter{}
	}
	p := pagination.NewOptions(opts...)

	q := r.sb.
		Select(usersTableColumns...).
		From(usersTable)

	// --- Фильтры ---
	if len(f.IDs) > 0 {
		q = q.Where(squirrel.Eq{"id": f.IDs})
	}
	if v := f.Email; v != nil && strings.TrimSpace(*v) != "" {
		q = q.Where("lower(email) = lower(?)", strings.TrimSpace(*v))
	}
	if v := f.Username; v != nil && *v != "" {
		q = q.Where("username ILIKE ?", "%"+*v+"%")
	}
	if v := f.FullText; v != nil && *v != "" {
		q = q.Where(
			"to_tsvector('simple', coalesce(full_name,'')) @@ plainto_tsquery('simple', ?)",
			*v,
		)
	}
	if v := f.IsActive; v != nil {
		q = q.Where(squirrel.Eq{"is_active": *v})
	}
	if v := f.CreatedFrom; v != nil {
		q = q.Where(squirrel.GtOrEq{"created_at": *v})
	}
	if v := f.CreatedTo; v != nil {
		q = q.Where(squirrel.Lt{"created_at": *v})
	}

	// --- Сортировка (whitelist) ---
	orderSQL := buildOrderBy(p.OrderBy())
	if len(orderSQL) == 0 {
		orderSQL = []string{"created_at DESC"}
	}
	q = q.OrderBy(orderSQL...)

	// --- Пагинация ---
	if l := p.Limit(); l > 0 {
		q = q.Limit(uint64(l))
	}
	if o := p.Offset(); o > 0 {
		q = q.Offset(uint64(o))
	}

	var rows []UserRow
	if err := r.pool.Selectx(ctx, &rows, q); err != nil {
		return nil, err
	}

	out := make([]*models.User, 0, len(rows))
	for i := range rows {
		out = append(out, ToModel(&rows[i]))
	}
	return out, nil
}

// --- helpers ---

// Разрешённые поля сортировки → SQL
func buildOrderBy(fields []pagination.SortField) []string {
	if len(fields) == 0 {
		return nil
	}

	whitelist := map[string]bool{
		"created_at": true,
		"username":   true,
		"email":      true,
		"last_login": true,
	}

	var out []string
	for _, f := range fields {
		name := strings.ToLower(strings.TrimSpace(f.Name))
		if !whitelist[name] {
			continue
		}
		dir := "ASC"
		if f.Desc {
			dir = "DESC"
		}
		out = append(out, fmt.Sprintf("%s %s", name, dir))
	}
	return out
}
