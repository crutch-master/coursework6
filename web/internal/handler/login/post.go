package login

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/crutch-master/coursework6/web/internal/auth"
	"github.com/crutch-master/coursework6/web/internal/data"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	"github.com/crutch-master/coursework6/web/internal/repository/user"
	"golang.org/x/crypto/bcrypt"
)

type PostHandler struct {
	templ    *template.Template
	userRepo *user.Repository
	secret   string
}

func NewPostHandler(templ *template.Template, userRepo *user.Repository, secret string) *PostHandler {
	return &PostHandler{
		templ:    templ,
		userRepo: userRepo,
		secret:   secret,
	}
}

func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	isAuth := middleware.IsAuthenticated(r.Context())

	login := r.FormValue("login")
	password := r.FormValue("password")

	u, err := h.userRepo.GetUserByLogin(r.Context(), login)
	if err != nil {
		slog.Error("failed to get user", "err", err)
		h.renderError(w, isAuth, "invalid name or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)); err != nil {
		h.renderError(w, isAuth, "invalid name or password")
		return
	}

	token, err := auth.CreateToken(u.ID, h.secret)
	if err != nil {
		slog.Error("failed to create token", "err", err)
		h.renderError(w, isAuth, "something went wrong")
		return
	}

	auth.SetAuthCookie(w, token)
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *PostHandler) renderError(w http.ResponseWriter, isAuth bool, msg string) {
	data := data.TemplateData{
		IsAuthenticated: isAuth,
		Error:           msg,
	}
	if err := h.templ.ExecuteTemplate(w, "base", data); err != nil {
		slog.Error("failed to execute template", "err", err)
	}
}

var _ http.Handler = (*PostHandler)(nil)

