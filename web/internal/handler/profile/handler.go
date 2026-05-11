package profile

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

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

func (h *Handler) View(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	u, err := h.userRepo.GetUserByID(r.Context(), id)
	if err != nil {
		slog.Error("failed to get user", "err", err)
		http.NotFound(w, r)
		return
	}

	currentUserID := middleware.GetUserID(r.Context())
	isAuth := middleware.IsAuthenticated(r.Context())

	d := data.TemplateData{
		IsAuthenticated: isAuth,
		Name:            u.Name,
		Description:     u.Description,
		ProfileID:       u.ID,
		IsOwner:         isAuth && currentUserID == u.ID,
	}

	if err := h.templ.ExecuteTemplate(w, "base", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

func (h *Handler) EditDescription(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		slog.Error("failed to parse form", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")

	if err := h.userRepo.UpdateDescription(r.Context(), userID, description); err != nil {
		slog.Error("failed to update description", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/user/"+strconv.FormatUint(userID, 10), http.StatusSeeOther)
}
