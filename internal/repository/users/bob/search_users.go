package users

import (
	"context"
	"strings"

	"github.com/moguchev/msa_open_lesson_postgres/internal/gen/bob/schema"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models/pagination"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/mods"
)

// query builder

func (r *Repository) SearchUsers(ctx context.Context, filter *models.UserFilter, opts ...pagination.Option) ([]*models.User, error) {
	query := schema.Users.Query(sm.Where(psql.And(toExpressions(filter)...)))

	// пагинация
	p := pagination.NewOptions(opts...)
	if limit := p.Limit(); limit > 0 {
		query.Apply(sm.Limit(limit))
	}
	if offset := p.Offset(); offset > 0 {
		query.Apply(sm.Offset(offset))
	}

	// сортировка (только whitelisted поля)
	for _, f := range p.OrderBy() {
		var expr bob.Expression
		switch f.Name {
		case "created_at":
			expr = schema.Users.Columns.CreatedAt
		case "username":
			expr = schema.Users.Columns.Username
		case "email":
			expr = schema.Users.Columns.Email
		case "last_login":
			expr = schema.Users.Columns.LastLogin
		default:
			continue
		}
		if f.Desc {
			query.Apply(sm.OrderBy(psql.Arg(expr)).Desc())
		} else {
			query.Apply(sm.OrderBy(psql.Arg(expr)).Asc())
		}
	}

	users, err := query.All(ctx, r.bobDB)
	if err != nil {
		return nil, err
	}

	result := make([]*models.User, 0, len(users))
	for _, u := range users {
		result = append(result, userFromSchema(u))
	}
	return result, nil
}

func toExpressions(filter *models.UserFilter) []bob.Expression {
	if filter == nil {
		return nil
	}

	var exprs []bob.Expression
	if v := filter.IDs; len(v) > 0 {
		exprs = append(exprs, schema.SelectWhere.Users.ID.In(v...).E)
	}

	if v := filter.Email; v != nil {
		exprs = append(exprs, schema.SelectWhere.Users.Email.EQ(strings.ToLower(*v)).E)
	}

	if v := filter.Username; v != nil && *v != "" {
		exprs = append(exprs, schema.SelectWhere.Users.Username.ILike("%"+*v+"%").E)
	}

	if v := filter.FullText; v != nil {
		expr := psql.Raw("to_tsvector('simple', coalesce(full_name, '')) @@ plainto_tsquery('simple', ?)", *v)
		exprs = append(exprs, mods.Where[*dialect.SelectQuery]{E: expr}.E)
	}

	if v := filter.CreatedFrom; v != nil {
		exprs = append(exprs, schema.SelectWhere.Users.CreatedAt.GTE(*v).E)
	}

	if v := filter.CreatedTo; v != nil {
		exprs = append(exprs, schema.SelectWhere.Users.CreatedAt.LTE(*v).E)
	}

	if v := filter.IsActive; v != nil {
		exprs = append(exprs, schema.SelectWhere.Users.IsActive.EQ(*v).E)
	}

	return exprs
}
