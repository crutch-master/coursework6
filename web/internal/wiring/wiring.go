package wiring

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/crutch-master/coursework6/web/internal/auth"
	"github.com/crutch-master/coursework6/web/internal/handler/article"
	"github.com/crutch-master/coursework6/web/internal/handler/index"
	"github.com/crutch-master/coursework6/web/internal/handler/login"
	"github.com/crutch-master/coursework6/web/internal/handler/profile"
	"github.com/crutch-master/coursework6/web/internal/handler/register"
	"github.com/crutch-master/coursework6/web/internal/handler/submit"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	articleRepo "github.com/crutch-master/coursework6/web/internal/repository/article"
	"github.com/crutch-master/coursework6/web/internal/repository/user"
	s3client "github.com/crutch-master/coursework6/web/internal/s3"
	"github.com/crutch-master/coursework6/web/static"
	"github.com/crutch-master/coursework6/web/templates"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func parsePage(base *template.Template, page string) *template.Template {
	return template.Must(template.Must(base.Clone()).ParseFS(templates.Templates, page))
}

func Wire(ctx context.Context) (http.Handler, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("godotenv.Load: %w", err)
	}

	pool, err := pgxpool.New(ctx, os.Getenv("DBSTRING"))
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	base := template.Must(template.ParseFS(templates.Templates, "base.html"))

	indexTempl := parsePage(base, "index.html")
	registerTempl := parsePage(base, "register.html")
	loginTempl := parsePage(base, "login.html")
	profileTempl := parsePage(base, "profile.html")
	submitTempl := parsePage(base, "submit.html")
	articleTempl := parsePage(base, "article.html")

	userRepo := user.NewRepository(pool)
	artRepo := articleRepo.NewRepository(pool)
	secret := os.Getenv("JWT_SECRET")
	s3c, err := s3client.NewClient(ctx,
		os.Getenv("S3_ENDPOINT"),
		os.Getenv("S3_REGION"),
		os.Getenv("S3_ACCESS_KEY"),
		os.Getenv("S3_SECRET_KEY"),
		os.Getenv("S3_BUCKET"),
	)
	if err != nil {
		return nil, fmt.Errorf("s3client.NewClient: %w", err)
	}

	mux := &http.ServeMux{}

	articleHandler := article.NewHandler(articleTempl, artRepo, s3c)

	mux.Handle("GET /", index.NewHandler(indexTempl))
	mux.Handle("GET /register", register.NewGetHandler(registerTempl))
	mux.Handle("POST /register", register.NewPostHandler(registerTempl, userRepo, secret))
	mux.Handle("GET /login", login.NewGetHandler(loginTempl))
	mux.Handle("POST /login", login.NewPostHandler(loginTempl, userRepo, secret))
	mux.Handle("GET /profile", middleware.RequireAuth(profile.NewHandler(profileTempl, userRepo)))
	mux.Handle("GET /submit", middleware.RequireAuth(submit.NewGetHandler(submitTempl)))
	mux.Handle("POST /submit", middleware.RequireAuth(submit.NewPostHandler(submitTempl, artRepo, s3c)))
	mux.Handle("GET /article/{id}", articleHandler)
	mux.HandleFunc("GET /article/{id}/source", articleHandler.Source)
	mux.HandleFunc("GET /article/{id}/pdf", articleHandler.PDF)
	mux.HandleFunc("GET /logout", func(w http.ResponseWriter, r *http.Request) {
		auth.ClearAuthCookie(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Static))))

	return middleware.WithAuth(secret, mux), nil
}