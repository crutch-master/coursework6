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

	columnID           = "id"
	columnLogin        = "login"
	columnName         = "name"
	columnDescription  = "description"
	columnPasswordHash = "password_hash"
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
		Select(columnID, columnLogin, columnName, columnDescription).
		From(table).
		Where(squirrel.Eq{columnID: id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return model.User{}, fmt.Errorf("squirrel.ToSql: %w", err)
	}

	var user model.User
	err = r.conn.QueryRow(ctx, sql, args...).Scan(&user.ID, &user.Login, &user.Name, &user.Description)
	if err != nil {
		return model.User{}, fmt.Errorf("row.Scan: %w", err)
	}

	return user, nil
}

func (r *Repository) CreateUser(ctx context.Context, login, name string, description string, passwordHash []byte) (uint64, error) {
	sql, args, err := squirrel.
		Insert(table).
		Columns(columnLogin, columnName, columnDescription, columnPasswordHash).
		Values(login, name, description, passwordHash).
		Suffix(fmt.Sprintf("RETURNING %s", columnID)).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return 0, fmt.Errorf("squirrel.ToSql: %w", err)
	}

	var id uint64
	err = r.conn.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("row.Scan: %w", err)
	}

	return id, nil
}

func (r *Repository) GetUserByLogin(ctx context.Context, login string) (model.User, error) {
	sql, args, err := squirrel.
		Select(columnID, columnLogin, columnName, columnDescription, columnPasswordHash).
		From(table).
		Where(squirrel.Eq{columnLogin: login}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return model.User{}, fmt.Errorf("squirrel.ToSql: %w", err)
	}

	var user model.User
	err = r.conn.QueryRow(ctx, sql, args...).Scan(&user.ID, &user.Login, &user.Name, &user.Description, &user.PasswordHash)
	if err != nil {
		return model.User{}, fmt.Errorf("row.Scan: %w", err)
	}

	return user, nil
}
