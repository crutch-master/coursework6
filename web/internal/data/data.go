package data

type TemplateData struct {
	IsAuthenticated bool
	Error           string
	Name            string
	Description     string
	DocumentName    string
	Status          string
	IsAuthor        bool
	ArticleID       uint64
}