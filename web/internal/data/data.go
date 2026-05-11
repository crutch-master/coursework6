package data

import "github.com/crutch-master/coursework6/web/internal/model"

type TemplateData struct {
	IsAuthenticated bool
	Error           string
	Name            string
	Description     string
	DocumentName    string
	Status          string
	IsAuthor        bool
	ArticleID       uint64
	ProfileID       uint64
	IsOwner         bool
	AuthorID        uint64
	AuthorName      string
	Articles        []model.ArticleListItem
}