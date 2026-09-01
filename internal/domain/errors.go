package domain

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserInactive    = errors.New("user inactive")
	ErrUserExists      = errors.New("user already exists")
	ErrSessionNotFound = errors.New("session not found")
	ErrDatabase        = errors.New("database error")
)
