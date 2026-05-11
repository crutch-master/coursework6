package profile

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/crutch-master/coursework6/web/internal/data"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	"github.com/crutch-master/coursework6/web/internal/repository/user"
)

type Handler struct {
	templ    *template.Template
	userRepo *user.Repository
}

func NewHandler(templ *template.Template, userRepo *user.Repository) *Handler {
	return &Handler{
		templ:    templ,
		userRepo: userRepo,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	u, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		slog.Error("failed to get user", "err", err)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	data := data.TemplateData{
		IsAuthenticated: true,
		Name:            u.Name,
		Description:     u.Description,
	}

	if err := h.templ.ExecuteTemplate(w, "base", data); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

var _ http.Handler = (*Handler)(nil)