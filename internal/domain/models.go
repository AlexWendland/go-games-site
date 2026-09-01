package domain

import (
	"time"
)

type User struct {
	UserId      string
	DisplayName string
	CreatedAt   time.Time
	IsActive    bool
}

type Session struct {
	Token     string
	CreatedAt time.Time
	ExpiresAt time.Time
}
