package submit

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/crutch-master/coursework6/web/internal/data"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	"github.com/crutch-master/coursework6/web/internal/repository/article"
	s3client "github.com/crutch-master/coursework6/web/internal/s3"
)

const maxUploadSize = 1 << 20

type PostHandler struct {
	templ       *template.Template
	articleRepo *article.Repository
	s3Client    *s3client.Client
}

func NewPostHandler(templ *template.Template, articleRepo *article.Repository, s3Client *s3client.Client) *PostHandler {
	return &PostHandler{
		templ:       templ,
		articleRepo: articleRepo,
		s3Client:    s3Client,
	}
}

func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	isAuth := middleware.IsAuthenticated(r.Context())
	userID := middleware.GetUserID(r.Context())

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.renderError(w, isAuth, "file exceeds 1MB limit")
		return
	}

	name := r.FormValue("name")
	if name == "" {
		h.renderError(w, isAuth, "document name is required")
		return
	}

	file, handler, err := r.FormFile("document")
	if err != nil {
		h.renderError(w, isAuth, "document is required")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(handler.Filename, ".typ") {
		h.renderError(w, isAuth, "only .typ files are accepted")
		return
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, io.LimitReader(file, maxUploadSize)); err != nil {
		slog.Error("failed to read document", "err", err)
		h.renderError(w, isAuth, "failed to read document")
		return
	}

	key, err := h.s3Client.Upload(r.Context(), handler.Filename, bytes.NewReader(buf.Bytes()))
	if err != nil {
		slog.Error("failed to upload document", "err", err)
		h.renderError(w, isAuth, "failed to upload document")
		return
	}

	id, err := h.articleRepo.CreateArticle(r.Context(), name, key, userID)
	if err != nil {
		slog.Error("failed to create article", "err", err)
		h.renderError(w, isAuth, "failed to create article")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/article/%d", id), http.StatusSeeOther)
}

func (h *PostHandler) renderError(w http.ResponseWriter, isAuth bool, msg string) {
	d := data.TemplateData{
		IsAuthenticated: isAuth,
		Error:           msg,
	}
	if err := h.templ.ExecuteTemplate(w, "base", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

var _ http.Handler = (*PostHandler)(nil)