package review

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/crutch-master/coursework6/web/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	table = "reviews"

	columnID         = "id"
	columnArticleID  = "article_id"
	columnReviewerID = "reviewer_id"
	columnText       = "text"
)

type Repository struct {
	conn *pgxpool.Pool
}

func NewRepository(conn *pgxpool.Pool) *Repository {
	return &Repository{
		conn: conn,
	}
}

func (r *Repository) CreateReview(ctx context.Context, articleID, reviewerID uint64, text string) (uint64, error) {
	sql, args, err := squirrel.
		Insert(table).
		Columns(columnArticleID, columnReviewerID, columnText).
		Values(articleID, reviewerID, text).
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

func (r *Repository) GetReviewByID(ctx context.Context, id uint64) (model.Review, error) {
	sql, args, err := squirrel.
		Select(columnID, columnArticleID, columnReviewerID, columnText).
		From(table).
		Where(squirrel.Eq{columnID: id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return model.Review{}, fmt.Errorf("squirrel.ToSql: %w", err)
	}

	var rev model.Review
	err = r.conn.QueryRow(ctx, sql, args...).Scan(&rev.ID, &rev.ArticleID, &rev.ReviewerID, &rev.Text)
	if err != nil {
		return model.Review{}, fmt.Errorf("row.Scan: %w", err)
	}

	return rev, nil
}

func (r *Repository) ListByArticle(ctx context.Context, articleID uint64) ([]model.ReviewListItem, error) {
	sql, args, err := squirrel.
		Select(
			"r."+columnID,
			"r."+columnReviewerID,
			"u.name",
		).
		From(table+" AS r").
		InnerJoin("users AS u ON r."+columnReviewerID+" = u.id").
		Where(squirrel.Eq{"r." + columnArticleID: articleID}).
		OrderBy("r." + columnID + " ASC").
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("squirrel.ToSql: %w", err)
	}

	rows, err := r.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("conn.Query: %w", err)
	}
	defer rows.Close()

	var items []model.ReviewListItem
	for rows.Next() {
		var item model.ReviewListItem
		if err := rows.Scan(&item.ID, &item.ReviewerID, &item.ReviewerName); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
