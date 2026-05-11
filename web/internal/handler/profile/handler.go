package profile

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/crutch-master/coursework6/web/internal/data"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	articleRepo "github.com/crutch-master/coursework6/web/internal/repository/article"
	reviewRepo "github.com/crutch-master/coursework6/web/internal/repository/review"
	"github.com/crutch-master/coursework6/web/internal/repository/user"
)

type Handler struct {
	templ         *template.Template
	editTempl     *template.Template
	partialTempl  *template.Template
	userRepo      *user.Repository
	articleRepo   *articleRepo.Repository
	reviewRepo    *reviewRepo.Repository
}

func NewHandler(templ *template.Template, editTempl *template.Template, partialTempl *template.Template, userRepo *user.Repository, articleRepo *articleRepo.Repository, reviewRepo *reviewRepo.Repository) *Handler {
	return &Handler{
		templ:         templ,
		editTempl:     editTempl,
		partialTempl:  partialTempl,
		userRepo:      userRepo,
		articleRepo:   articleRepo,
		reviewRepo:    reviewRepo,
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

	articles, err := h.articleRepo.ListByAuthor(r.Context(), u.ID)
	if err != nil {
		slog.Error("failed to list user articles", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	d := data.TemplateData{
		IsAuthenticated: isAuth,
		Name:            u.Name,
		Description:     u.Description,
		ProfileID:       u.ID,
		IsOwner:         isAuth && currentUserID == u.ID,
		UserArticles:    articles,
	}

	if err := h.templ.ExecuteTemplate(w, "base", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

func (h *Handler) EditPage(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	currentUserID := middleware.GetUserID(r.Context())
	if currentUserID != id {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	u, err := h.userRepo.GetUserByID(r.Context(), id)
	if err != nil {
		slog.Error("failed to get user", "err", err)
		http.NotFound(w, r)
		return
	}

	d := data.TemplateData{
		IsAuthenticated: true,
		Name:            u.Name,
		Description:     u.Description,
		ProfileID:       u.ID,
		IsOwner:         true,
	}

	if err := h.editTempl.ExecuteTemplate(w, "base", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

func (h *Handler) EditProfile(w http.ResponseWriter, r *http.Request) {
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

	name := r.FormValue("name")
	description := r.FormValue("description")

	if name == "" {
		slog.Error("name is required")
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if err := h.userRepo.UpdateProfile(r.Context(), userID, name, description); err != nil {
		slog.Error("failed to update profile", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/user/"+strconv.FormatUint(userID, 10), http.StatusSeeOther)
}

func (h *Handler) ArticlesTab(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	articles, err := h.articleRepo.ListByAuthor(r.Context(), id)
	if err != nil {
		slog.Error("failed to list user articles", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	d := data.TemplateData{
		UserArticles: articles,
	}

	if err := h.partialTempl.ExecuteTemplate(w, "profile_articles_partial", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

func (h *Handler) ReviewsTab(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	reviews, err := h.reviewRepo.ListByReviewer(r.Context(), id)
	if err != nil {
		slog.Error("failed to list user reviews", "err", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	d := data.TemplateData{
		UserReviews: reviews,
	}

	if err := h.partialTempl.ExecuteTemplate(w, "profile_reviews_partial", d); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}
