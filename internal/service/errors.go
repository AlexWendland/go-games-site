package service

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserInactive = errors.New("user inactive")
	ErrUserExists   = errors.New("user already exists")
	ErrSessionNotFound = errors.New("session not found")
)
