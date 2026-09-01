package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/AlexWendland/go-games-site/internal/config"
	"github.com/AlexWendland/go-games-site/internal/domain"
	"github.com/AlexWendland/go-games-site/internal/middleware"
)

type AuthService interface {
	LogIn(ctx context.Context, userID string, createdAt time.Time) (*domain.User, *domain.Session, error)
	LogOut(ctx context.Context, token string) error
}

type UserService interface {
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	CreateUser(ctx context.Context, userID string, displayName string, createdAt time.Time) (*domain.User, error)
	UpdateDisplayName(ctx context.Context, userID string, displayName string) (*domain.User, error)
}

type Handler struct {
	cfg         *config.Config
	authService AuthService
	userService UserService
}

func toUserResponse(u *domain.User) UserResponse {
	return UserResponse{
		UserId:      &u.UserId,
		DisplayName: &u.DisplayName,
		CreatedAt:   &u.CreatedAt,
		IsActive:    &u.IsActive,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h Handler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req LogInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, session, err := h.authService.LogIn(r.Context(), req.UserId, time.Now())
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			http.Error(w, "unauthorised", http.StatusUnauthorized)
			return
		}
		if errors.Is(err, domain.ErrDatabase) {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     domain.SessionCookieName,
		Value:    session.Token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.Production,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     domain.UserIDCookieName,
		Value:    req.UserId,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.Production,
		Path:     "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:     domain.SessionExpriationCookieName,
		Value:    session.ExpiresAt.String(),
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   h.cfg.Production,
		Path:     "/",
	})
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h Handler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	token, ok := middleware.TokenFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return
	}
	if err := h.authService.LogOut(r.Context(), token); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	user, err := h.userService.GetUser(r.Context(), userId)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.userService.CreateUser(r.Context(), req.UserId, req.DisplayName, time.Now())
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
	}
	writeJSON(w, http.StatusCreated, toUserResponse(user))
}

func (h Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return
	}
	newUser, err := h.userService.UpdateDisplayName(r.Context(), user.UserId, req.DisplayName)
	if err != nil {
		http.Error(w, "user did not exist", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(newUser))
}

func (h Handler) GetSession(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "user did not exist", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
}

func ApiHandler(cfg *config.Config, authService AuthService, userService UserService, authMiddleware func(http.Handler) http.Handler) http.Handler {
	apiRouter := http.NewServeMux()
	h := Handler{cfg, authService, userService}

	// Docs page
	apiRouter.Handle("/docs/", http.StripPrefix("/docs", swaggerUIHandler()))

	// Auth routes
	apiRouter.HandleFunc("POST /auth", h.CreateSession)
	apiRouter.Handle("DELETE /auth", authMiddleware(http.HandlerFunc(h.DeleteSession)))
	apiRouter.Handle("GET /auth", authMiddleware(http.HandlerFunc(h.GetSession)))

	// User routes
	apiRouter.HandleFunc("GET /user", h.GetUser)
	apiRouter.HandleFunc("POST /user", h.CreateUser)
	apiRouter.Handle("PUT /user", authMiddleware(http.HandlerFunc(h.UpdateUser)))

	return apiRouter
}
