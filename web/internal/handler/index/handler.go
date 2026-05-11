package index

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
)

type Handler struct {
	templ *template.Template
}

func NewHandler(templ *template.Template) *Handler {
	return &Handler{
		templ: templ,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.templ.Execute(w, struct{}{}); err != nil {
		slog.Error("failed to execute template", "err", fmt.Errorf("h.templ.Execute: %w", err))
	}
}

var _ http.Handler = (*Handler)(nil)
