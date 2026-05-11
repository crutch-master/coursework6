package review

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/crutch-master/coursework6/web/internal/data"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	articleRepo "github.com/crutch-master/coursework6/web/internal/repository/article"
)

type GetHandler struct {
	templ       *template.Template
	articleRepo *articleRepo.Repository
}

func NewGetHandler(templ *template.Template, articleRepo *articleRepo.Repository) *GetHandler {
	return &GetHandler{
		templ:       templ,
		articleRepo: articleRepo,
	}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	d := data.TemplateData{
		IsAuthenticated: middleware.IsAuthenticated(r.Context()),
		ArticleID:       a.ID,
		DocumentName:    a.DocumentName,
	}

	if err := h.templ.ExecuteTemplate(w, "base", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

var _ http.Handler = (*GetHandler)(nil)
