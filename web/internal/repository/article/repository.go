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