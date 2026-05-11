package article

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/crutch-master/coursework6/web/internal/data"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	"github.com/crutch-master/coursework6/web/internal/model"
	articleRepo "github.com/crutch-master/coursework6/web/internal/repository/article"
	s3client "github.com/crutch-master/coursework6/web/internal/s3"
)

type Handler struct {
	templ       *template.Template
	articleRepo *articleRepo.Repository
	s3Client    *s3client.Client
}

func NewHandler(templ *template.Template, articleRepo *articleRepo.Repository, s3Client *s3client.Client) *Handler {
	return &Handler{
		templ:       templ,
		articleRepo: articleRepo,
		s3Client:    s3Client,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	a, err := h.articleRepo.GetArticleByID(r.Context(), id)
	if err != nil {
		slog.Error("failed to get article", "err", err)
		http.NotFound(w, r)
		return
	}

	if a.Status == "pending" || a.Status == "failed" {
		userID := middleware.GetUserID(r.Context())
		if userID != a.AuthorID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	isAuth := middleware.IsAuthenticated(r.Context())

	d := data.TemplateData{
		IsAuthenticated: isAuth,
		DocumentName:    a.DocumentName,
		Status:          a.Status,
		IsAuthor:        isAuth && middleware.GetUserID(r.Context()) == a.AuthorID,
		ArticleID:       a.ID,
	}

	if err := h.templ.ExecuteTemplate(w, "base", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

func (h *Handler) Source(w http.ResponseWriter, r *http.Request) {
	a, ok := h.getArticleWithAccess(w, r)
	if !ok {
		return
	}

	out, err := h.s3Client.Download(r.Context(), a.DocumentFID)
	if err != nil {
		slog.Error("failed to download source", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer out.Body.Close()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.typ"`, a.DocumentName))
	io.Copy(w, out.Body)
}

func (h *Handler) PDF(w http.ResponseWriter, r *http.Request) {
	a, ok := h.getArticleWithAccess(w, r)
	if !ok {
		return
	}

	if a.PDFID == nil {
		http.NotFound(w, r)
		return
	}

	out, err := h.s3Client.Download(r.Context(), *a.PDFID)
	if err != nil {
		slog.Error("failed to download pdf", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer out.Body.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s.pdf"`, a.DocumentName))
	io.Copy(w, out.Body)
}

func (h *Handler) getArticleWithAccess(w http.ResponseWriter, r *http.Request) (*model.Article, bool) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return nil, false
	}

	a, err := h.articleRepo.GetArticleByID(r.Context(), id)
	if err != nil {
		slog.Error("failed to get article", "err", err)
		http.NotFound(w, r)
		return nil, false
	}

	if a.Status == "pending" || a.Status == "failed" {
		userID := middleware.GetUserID(r.Context())
		if userID != a.AuthorID {
			http.Error(w, "forbidden", http.StatusForbidden)
			return nil, false
		}
	}

	return &a, true
}

var _ http.Handler = (*Handler)(nil)