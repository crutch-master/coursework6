package model

type Review struct {
	ID         uint64
	ArticleID  uint64
	ReviewerID uint64
	Text       string
}

type ReviewListItem struct {
	ID           uint64
	ReviewerID   uint64
	ReviewerName string
}
