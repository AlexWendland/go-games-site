package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/AlexWendland/go-games-site/internal/domain"
)

type contextKey string

const UserContextKey contextKey = "user"
const TokenContextKey contextKey = "token"

type AuthService interface {
	GetUserBySession(ctx context.Context, token string, currentTime time.Time) (*domain.User, error)
}

func Auth(authService AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from request
			token, err := r.Cookie(domain.SessionCookieName)
			if err != nil {
				http.Error(w, "unauthorised", http.StatusUnauthorized)
				return
			}

			// Validate token against database
			user, err := authService.GetUserBySession(r.Context(), token.Value, time.Now())
			if err != nil {
				http.Error(w, "unauthorised", http.StatusUnauthorized)
				return
			}

			// Attach user and token to context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			ctx = context.WithValue(ctx, TokenContextKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(ctx context.Context) (*domain.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*domain.User)
	return user, ok
}

func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(TokenContextKey).(string)
	return token, ok
}
