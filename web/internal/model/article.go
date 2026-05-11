package model

type Article struct {
	ID           uint64
	DocumentName string
	DocumentFID  string
	AuthorID     uint64
	Status       string
	PDFID        *string
}

type ArticleListItem struct {
	ID           uint64
	DocumentName string
	AuthorID     uint64
	AuthorName   string
}