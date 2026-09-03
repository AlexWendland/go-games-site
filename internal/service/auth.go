package service

import (
	"context"
	"crypto/rand"
	"net/http"
	"time"

	"github.com/AlexWendland/go-games-site/internal/db"
	"github.com/AlexWendland/go-games-site/internal/domain"
)

type contextKey string

const (
	DayInNanoSeconds            = 86_400_000_000_000
	userContextKey   contextKey = "user"
	tokenContextKey  contextKey = "token"
)

type AuthService struct {
	db *db.DB
}

func MakeAuthService(db *db.DB) *AuthService {
	return &AuthService{db}
}

func (as AuthService) LogIn(ctx context.Context, userID string, createdAt time.Time) (*domain.User, *domain.Session, error) {
	var user db.User
	var session db.Session

	// Wrap db calls in a single transaction
	err := as.db.WithTx(ctx, func(q *db.Queries) error {
		var err error
		user, err = q.GetActiveUserByUserID(ctx, userID)
		if err != nil {
			return domain.ErrUserNotFound
		}
		token := rand.Text()
		session, err = q.CreateSession(ctx, db.CreateSessionParams{
			UserID:       user.ID,
			SessionToken: token,
			CreatedAt:    createdAt,
			ExpiresAt:    createdAt.Add(DayInNanoSeconds),
		})
		if err != nil {
			return domain.ErrDatabase
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return toUser(user), &domain.Session{Token: session.SessionToken, CreatedAt: session.CreatedAt, ExpiresAt: session.ExpiresAt}, nil
}

func (as AuthService) LogOut(ctx context.Context, token string) error {
	if as.db.DeleteSessionByToken(ctx, token) != nil {
		return domain.ErrDatabase
	}
	return nil
}

func (as AuthService) GetUserBySession(ctx context.Context, token string, currentTime time.Time) (*domain.User, error) {
	user, err := as.db.GetActiveUserBySessionToken(ctx, db.GetActiveUserBySessionTokenParams{SessionToken: token, ExpiresAt: currentTime})
	if err != nil {
		return nil, domain.ErrSessionNotFound
	}
	return toUser(user), nil
}

func (as AuthService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from request
		token, err := r.Cookie(domain.SessionCookieName)
		if err != nil {
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}

		// Validate token against database
		user, err := as.GetUserBySession(r.Context(), token.Value, time.Now())
		if err != nil {
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}

		// Attach user and token to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		ctx = context.WithValue(ctx, tokenContextKey, token.Value)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (as AuthService) UserFromRequest(r *http.Request) (*domain.User, bool) {
	user, ok := r.Context().Value(userContextKey).(*domain.User)
	return user, ok
}

func (as AuthService) TokenFromRequest(r *http.Request) (string, bool) {
	token, ok := r.Context().Value(tokenContextKey).(string)
	return token, ok
}
