package review

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/crutch-master/coursework6/web/internal/data"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	articleRepo "github.com/crutch-master/coursework6/web/internal/repository/article"
	reviewRepo "github.com/crutch-master/coursework6/web/internal/repository/review"
)

type PostHandler struct {
	templ       *template.Template
	reviewRepo  *reviewRepo.Repository
	articleRepo *articleRepo.Repository
}

func NewPostHandler(templ *template.Template, reviewRepo *reviewRepo.Repository, articleRepo *articleRepo.Repository) *PostHandler {
	return &PostHandler{
		templ:       templ,
		reviewRepo:  reviewRepo,
		articleRepo: articleRepo,
	}
}

func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	isAuth := middleware.IsAuthenticated(r.Context())
	userID := middleware.GetUserID(r.Context())

	articleIDStr := r.PathValue("id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	a, err := h.articleRepo.GetArticleByID(r.Context(), articleID)
	if err != nil {
		slog.Error("failed to get article", "err", err)
		http.NotFound(w, r)
		return
	}

	if a.Status != "published" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse form", "err", err)
		h.renderError(w, isAuth, articleID, a.DocumentName, "failed to parse form")
		return
	}

	text := r.FormValue("text")
	if text == "" {
		h.renderError(w, isAuth, articleID, a.DocumentName, "review text is required")
		return
	}

	id, err := h.reviewRepo.CreateReview(r.Context(), articleID, userID, text)
	if err != nil {
		slog.Error("failed to create review", "err", err)
		h.renderError(w, isAuth, articleID, a.DocumentName, "failed to create review")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/review/%d", id), http.StatusSeeOther)
}

func (h *PostHandler) renderError(w http.ResponseWriter, isAuth bool, articleID uint64, documentName, msg string) {
	d := data.TemplateData{
		IsAuthenticated: isAuth,
		ArticleID:       articleID,
		DocumentName:    documentName,
		Error:           msg,
	}
	if err := h.templ.ExecuteTemplate(w, "base", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

var _ http.Handler = (*PostHandler)(nil)
