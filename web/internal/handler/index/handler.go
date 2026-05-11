package index

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/crutch-master/coursework6/web/internal/data"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	articleRepo "github.com/crutch-master/coursework6/web/internal/repository/article"
)

type Handler struct {
	templ       *template.Template
	articleRepo *articleRepo.Repository
	latestCount int
}

func NewHandler(templ *template.Template, articleRepo *articleRepo.Repository, latestCount int) *Handler {
	return &Handler{
		templ:       templ,
		articleRepo: articleRepo,
		latestCount: latestCount,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	articles, err := h.articleRepo.ListPublished(r.Context(), h.latestCount)
	if err != nil {
		slog.Error("failed to list published articles", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	d := data.TemplateData{
		IsAuthenticated: middleware.IsAuthenticated(r.Context()),
		Articles:        articles,
	}

	if err := h.templ.ExecuteTemplate(w, "base", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

func (h *Handler) ProfileRedirect(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/user/"+strconv.FormatUint(userID, 10), http.StatusSeeOther)
}

var _ http.Handler = (*Handler)(nil)
