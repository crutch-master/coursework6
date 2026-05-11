package login

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/crutch-master/coursework6/web/internal/data"
	"github.com/crutch-master/coursework6/web/internal/middleware"
)

type GetHandler struct {
	templ *template.Template
}

func NewGetHandler(templ *template.Template) *GetHandler {
	return &GetHandler{templ: templ}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := data.TemplateData{
		IsAuthenticated: middleware.IsAuthenticated(r.Context()),
	}
	if err := h.templ.ExecuteTemplate(w, "base", data); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

var _ http.Handler = (*GetHandler)(nil)