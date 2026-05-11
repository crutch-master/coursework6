package wiring

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/crutch-master/coursework6/web/internal/handler/index"
	"github.com/crutch-master/coursework6/web/internal/repository/user"
	"github.com/crutch-master/coursework6/web/static"
	"github.com/crutch-master/coursework6/web/templates"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func Wire(ctx context.Context) (http.Handler, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("godotenv.Load: %w", err)
	}

	pool, err := pgxpool.New(ctx, os.Getenv("DBSTRING"))
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New: %w", err)
	}

	templ, err := template.ParseFS(templates.Templates, "*.html")
	if err != nil {
		return nil, fmt.Errorf("template.ParseFS: %w", err)
	}

	userRepo := user.NewRepository(pool)

	handler := &http.ServeMux{}

	handler.Handle("GET /", index.NewHandler(templ))

	handler.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.Static))))

	_ = userRepo

	return handler, nil
}
