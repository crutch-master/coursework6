package user

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/crutch-master/coursework6/web/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	table = "users"

	columnID          = "id"
	columnName        = "name"
	columnDescription = "description"
)

type Repository struct {
	conn *pgxpool.Pool
}

func NewRepository(conn *pgxpool.Pool) *Repository {
	return &Repository{
		conn: conn,
	}
}

func (r *Repository) GetUserByID(ctx context.Context, id uint64) (model.User, error) {
	sql, args, err := squirrel.
		Select(columnID, columnName, columnDescription).
		From(table).
		Where(squirrel.Eq{columnID: id}).
		ToSql()

	if err != nil {
		return model.User{}, fmt.Errorf("squirrel.ToSql: %w", err)
	}

	var user model.User
	err = r.conn.QueryRow(ctx, sql, args...).Scan(&user.ID, &user.Name, &user.Description)
	if err != nil {
		return model.User{}, fmt.Errorf("row.Scan: %w", err)
	}

	return user, nil
}
