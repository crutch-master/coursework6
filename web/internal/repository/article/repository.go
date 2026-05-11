package article

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/crutch-master/coursework6/web/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	table = "articles"

	columnID           = "id"
	columnDocumentName = "document_name"
	columnDocumentFID  = "document_fid"
	columnAuthorID     = "author_id"
	columnStatus       = "status"
	columnPDFID        = "pdf_fid"
)

type Repository struct {
	conn *pgxpool.Pool
}

func NewRepository(conn *pgxpool.Pool) *Repository {
	return &Repository{
		conn: conn,
	}
}

func (r *Repository) CreateArticle(ctx context.Context, documentName, documentFID string, authorID uint64) (uint64, error) {
	sql, args, err := squirrel.
		Insert(table).
		Columns(columnDocumentName, columnDocumentFID, columnAuthorID, columnStatus).
		Values(documentName, documentFID, authorID, "pending").
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

func (r *Repository) GetArticleByID(ctx context.Context, id uint64) (model.Article, error) {
	sql, args, err := squirrel.
		Select(columnID, columnDocumentName, columnDocumentFID, columnAuthorID, columnStatus, columnPDFID).
		From(table).
		Where(squirrel.Eq{columnID: id}).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return model.Article{}, fmt.Errorf("squirrel.ToSql: %w", err)
	}

	var a model.Article
	err = r.conn.QueryRow(ctx, sql, args...).Scan(&a.ID, &a.DocumentName, &a.DocumentFID, &a.AuthorID, &a.Status, &a.PDFID)
	if err != nil {
		return model.Article{}, fmt.Errorf("row.Scan: %w", err)
	}

	return a, nil
}

func (r *Repository) ListPublished(ctx context.Context, limit int) ([]model.ArticleListItem, error) {
	sql, args, err := squirrel.
		Select(
			"a."+columnID,
			"a."+columnDocumentName,
			"a."+columnAuthorID,
			"u.name",
		).
		From(table+" AS a").
		InnerJoin("users AS u ON a."+columnAuthorID+" = u.id").
		Where(squirrel.Eq{"a." + columnStatus: "published"}).
		OrderBy("a." + columnID + " DESC").
		Limit(uint64(limit)).
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

	var items []model.ArticleListItem
	for rows.Next() {
		var item model.ArticleListItem
		if err := rows.Scan(&item.ID, &item.DocumentName, &item.AuthorID, &item.AuthorName); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}