package service

import (
	"context"
	"github.com/AlexWendland/go-games-site/internal/db"
	"time"
)

type UserService struct {
	db *db.DB
}

func MakeUserService(db *db.DB) *UserService {
	return &UserService{db}
}

func toUser(user db.User) *User {
	return &User{
		UserId:      user.UserID,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		IsActive:    user.IsActive,
	}
}

func (us UserService) DoesUserExist(ctx context.Context, userID string) (*User, error) {
	user, err := us.db.GetUserByUserID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return toUser(user), nil
}

func (us UserService) CreateUser(ctx context.Context, userID string, displayName string, createdAt time.Time) (*User, error) {
	user, err := us.db.CreateUser(ctx, db.CreateUserParams{
		UserID: userID, DisplayName: displayName, CreatedAt: createdAt, IsActive: true,
	})
	if err != nil {
		return nil, ErrUserExists
	}
	return toUser(user), nil
}

func (us UserService) UpdateDisplayName(ctx context.Context, userID string, displayName string) (*User, error) {
	user, err := us.db.UpdateDisplayNameByUserID(ctx, db.UpdateDisplayNameByUserIDParams{
		DisplayName: displayName,
		UserID:      userID,
	})
	if err != nil {
		return nil, ErrUserNotFound
	}
	return toUser(user), nil
}
