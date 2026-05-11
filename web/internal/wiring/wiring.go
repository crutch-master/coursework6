package wiring

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/crutch-master/coursework6/web/internal/auth"
	"github.com/crutch-master/coursework6/web/internal/handler/article"
	"github.com/crutch-master/coursework6/web/internal/handler/index"
	"github.com/crutch-master/coursework6/web/internal/handler/login"
	"github.com/crutch-master/coursework6/web/internal/handler/profile"
	"github.com/crutch-master/coursework6/web/internal/handler/register"
	"github.com/crutch-master/coursework6/web/internal/handler/review"
	"github.com/crutch-master/coursework6/web/internal/handler/submit"
	"github.com/crutch-master/coursework6/web/internal/middleware"
	articleRepo "github.com/crutch-master/coursework6/web/internal/repository/article"
	reviewRepo "github.com/crutch-master/coursework6/web/internal/repository/review"
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

func getEnvInt(key string, defaultVal int) int {
	s := os.Getenv(key)
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return n
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
	editProfileTempl := parsePage(base, "edit_profile.html")
	profilePartials := template.Must(template.New("partials").ParseFS(templates.Templates, "profile_articles_partial.html", "profile_reviews_partial.html"))
	submitTempl := parsePage(base, "submit.html")
	articleTempl := parsePage(base, "article.html")
	reviewTempl := parsePage(base, "review.html")
	submitReviewTempl := parsePage(base, "submit_review.html")

	userRepo := user.NewRepository(pool)
	artRepo := articleRepo.NewRepository(pool)
	revRepo := reviewRepo.NewRepository(pool)
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

	latestCount := getEnvInt("INDEX_LATEST_COUNT", 5)
	indexHandler := index.NewHandler(indexTempl, artRepo, latestCount)
	profileHandler := profile.NewHandler(profileTempl, editProfileTempl, profilePartials, userRepo, artRepo, revRepo)
	articleHandler := article.NewHandler(articleTempl, artRepo, s3c, userRepo, revRepo)
	reviewHandler := review.NewHandler(reviewTempl, revRepo, artRepo, userRepo)

	mux := &http.ServeMux{}

	mux.Handle("GET /", indexHandler)
	mux.Handle("GET /register", register.NewGetHandler(registerTempl))
	mux.Handle("POST /register", register.NewPostHandler(registerTempl, userRepo, secret))
	mux.Handle("GET /login", login.NewGetHandler(loginTempl))
	mux.Handle("POST /login", login.NewPostHandler(loginTempl, userRepo, secret))
	mux.HandleFunc("GET /profile", indexHandler.ProfileRedirect)
	mux.Handle("POST /profile/edit", middleware.RequireAuth(http.HandlerFunc(profileHandler.EditProfile)))
	mux.Handle("GET /user/{id}", http.HandlerFunc(profileHandler.View))
	mux.Handle("GET /user/{id}/edit", middleware.RequireAuth(http.HandlerFunc(profileHandler.EditPage)))
	mux.HandleFunc("GET /user/{id}/tab/articles", profileHandler.ArticlesTab)
	mux.HandleFunc("GET /user/{id}/tab/reviews", profileHandler.ReviewsTab)
	mux.Handle("GET /submit", middleware.RequireAuth(submit.NewGetHandler(submitTempl)))
	mux.Handle("POST /submit", middleware.RequireAuth(submit.NewPostHandler(submitTempl, artRepo, s3c)))
	mux.Handle("GET /article/{id}", articleHandler)
	mux.HandleFunc("GET /article/{id}/source", articleHandler.Source)
	mux.HandleFunc("GET /article/{id}/pdf", articleHandler.PDF)
	mux.Handle("GET /article/{id}/review", middleware.RequireAuth(review.NewGetHandler(submitReviewTempl, artRepo)))
	mux.Handle("POST /article/{id}/review", middleware.RequireAuth(review.NewPostHandler(submitReviewTempl, revRepo, artRepo)))
	mux.HandleFunc("GET /review/{id}", reviewHandler.View)
	mux.HandleFunc("GET /logout", func(w http.ResponseWriter, r *http.Request) {
		auth.ClearAuthCookie(w)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Static))))

	return middleware.WithAuth(secret, mux), nil
}
