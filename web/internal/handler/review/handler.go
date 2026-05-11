package review

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/crutch-master/coursework6/web/internal/data"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	articleRepo "github.com/crutch-master/coursework6/web/internal/repository/article"
	reviewRepo "github.com/crutch-master/coursework6/web/internal/repository/review"
	userRepo "github.com/crutch-master/coursework6/web/internal/repository/user"
)

type Handler struct {
	templ       *template.Template
	reviewRepo  *reviewRepo.Repository
	articleRepo *articleRepo.Repository
	userRepo    *userRepo.Repository
}

func NewHandler(templ *template.Template, reviewRepo *reviewRepo.Repository, articleRepo *articleRepo.Repository, userRepo *userRepo.Repository) *Handler {
	return &Handler{
		templ:       templ,
		reviewRepo:  reviewRepo,
		articleRepo: articleRepo,
		userRepo:    userRepo,
	}
}

func (h *Handler) View(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rev, err := h.reviewRepo.GetReviewByID(r.Context(), id)
	if err != nil {
		slog.Error("failed to get review", "err", err)
		http.NotFound(w, r)
		return
	}

	a, err := h.articleRepo.GetArticleByID(r.Context(), rev.ArticleID)
	if err != nil {
		slog.Error("failed to get article", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	author, err := h.userRepo.GetUserByID(r.Context(), a.AuthorID)
	if err != nil {
		slog.Error("failed to get article author", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	reviewer, err := h.userRepo.GetUserByID(r.Context(), rev.ReviewerID)
	if err != nil {
		slog.Error("failed to get reviewer", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	d := data.TemplateData{
		IsAuthenticated: middleware.IsAuthenticated(r.Context()),
		ReviewID:        rev.ID,
		ReviewText:      rev.Text,
		ReviewerID:      rev.ReviewerID,
		ReviewerName:    reviewer.Name,
		ArticleID:       a.ID,
		DocumentName:    a.DocumentName,
		AuthorID:        a.AuthorID,
		AuthorName:      author.Name,
	}

	if err := h.templ.ExecuteTemplate(w, "base", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}
