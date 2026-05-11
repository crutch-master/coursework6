package data

import "github.com/crutch-master/coursework6/web/internal/model"

type TemplateData struct {
	IsAuthenticated    bool
	Error              string
	Name               string
	Description        string
	ArticleDescription string
	DocumentName       string
	Status             string
	IsAuthor           bool
	ArticleID          uint64
	ProfileID          uint64
	IsOwner            bool
	AuthorID           uint64
	AuthorName         string
	ReviewID           uint64
	ReviewText         string
	ReviewerID         uint64
	ReviewerName       string
	Articles           []model.ArticleListItem
	Reviews            []model.ReviewListItem
	UserArticles       []model.ArticleListItem
	UserReviews        []model.UserReviewItem
}