package service

import (
	"context"
	"crypto/rand"
	"errors"
	"github.com/AlexWendland/go-games-site/internal/db"
	"time"
)

const (
	SessionCookieName           = "session"
	UserIDCookieName            = "user_id"
	SessionExpriationCookieName = "session_expires_at"
	DayInNanoSeconds            = 86_400_000_000_000
)

type AuthService struct {
	db *db.DB
}

func MakeAuthService(db *db.DB) *AuthService {
	return &AuthService{db}
}

func (as AuthService) LogIn(ctx context.Context, userID string, createdAt time.Time) (*User, *Session, error) {
	var user db.User
	var session db.Session

	// Wrap db calls in a single transaction
	err := as.db.WithTx(ctx, func(q *db.Queries) error {
		var err error
		user, err = q.GetActiveUserByUserID(ctx, userID)
		if err != nil {
			return ErrUserNotFound
		}
		token := rand.Text()
		session, err = q.CreateSession(ctx, db.CreateSessionParams{
			UserID:       user.ID,
			SessionToken: token,
			CreatedAt:    createdAt,
			ExpiresAt:    createdAt.Add(DayInNanoSeconds),
		})
		if err != nil {
			return errors.New("could not create session")
		}
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return toUser(user), &Session{Token: session.SessionToken, CreatedAt: session.CreatedAt, ExpiresAt: session.ExpiresAt}, nil
}

func (as AuthService) LogOut(ctx context.Context, token string) error {
	return as.db.DeleteSessionByToken(ctx, token)
}

func (as AuthService) GetUserBySession(ctx context.Context, token string, currentTime time.Time) (*User, error) {
	user, err := as.db.GetActiveUserBySessionToken(ctx, db.GetActiveUserBySessionTokenParams{SessionToken: token, ExpiresAt: currentTime})
	if err != nil {
		return nil, err
	}
	return toUser(user), nil
}
