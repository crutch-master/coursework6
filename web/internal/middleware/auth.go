package middleware

import (
	"context"
	"net/http"

	"github.com/crutch-master/coursework6/web/internal/auth"
)

type ctxKey struct{}

func GetUserID(ctx context.Context) uint64 {
	v, _ := ctx.Value(ctxKey{}).(uint64)
	return v
}

func IsAuthenticated(ctx context.Context) bool {
	return GetUserID(ctx) > 0
}

func WithAuth(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := auth.GetUserIDFromRequest(r, secret)
		if err != nil {
			userID = 0
		}
		ctx := context.WithValue(r.Context(), ctxKey{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r.Context()) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}