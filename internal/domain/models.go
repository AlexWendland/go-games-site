package domain

import (
	"time"
)

type PlayerPositionNumber uint64

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

type PlayerId struct {
	UserId string
	IsAi   bool
}

type PlayerPosition struct {
	Position PlayerPositionNumber
	Player   PlayerId
}
