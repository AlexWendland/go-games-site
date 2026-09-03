package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AlexWendland/go-games-site/internal/config"
	"github.com/AlexWendland/go-games-site/internal/domain"

	"github.com/google/go-cmp/cmp"
)

const NO_COOKIE = "__NO_COOKIE__"

var (
	Date1 = time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	Date2 = time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)
)

type AuthServiceStub struct {
	logIn            func(ctx context.Context, userID string, createdAt time.Time) (*domain.User, *domain.Session, error)
	logOut           func(ctx context.Context, token string) error
	authMiddleware   func(http.Handler) http.Handler
	userFromRequest  func(r *http.Request) (*domain.User, bool)
	tokenFromRequest func(r *http.Request) (string, bool)
}

func (s AuthServiceStub) LogIn(ctx context.Context, userID string, createdAt time.Time) (*domain.User, *domain.Session, error) {
	return s.logIn(ctx, userID, createdAt)
}

func (s AuthServiceStub) LogOut(ctx context.Context, token string) error {
	return s.logOut(ctx, token)
}

func (s AuthServiceStub) AuthMiddleware(next http.Handler) http.Handler {
	return s.authMiddleware(next)
}

func (s AuthServiceStub) UserFromRequest(r *http.Request) (*domain.User, bool) {
	return s.userFromRequest(r)
}

func (s AuthServiceStub) TokenFromRequest(r *http.Request) (string, bool) {
	return s.tokenFromRequest(r)
}

type UserServiceStub struct {
	getUser           func(ctx context.Context, userID string) (*domain.User, error)
	createUser        func(ctx context.Context, userID string, displayName string, createdAt time.Time) (*domain.User, error)
	updateDisplayName func(ctx context.Context, userID string, displayName string) (*domain.User, error)
}

func (s UserServiceStub) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	return s.getUser(ctx, userID)
}

func (s UserServiceStub) CreateUser(ctx context.Context, userID string, displayName string, createdAt time.Time) (*domain.User, error) {
	return s.createUser(ctx, userID, displayName, createdAt)
}

func (s UserServiceStub) UpdateDisplayName(ctx context.Context, userID string, displayName string) (*domain.User, error) {
	return s.updateDisplayName(ctx, userID, displayName)
}

func checkCookie(t *testing.T, w *httptest.ResponseRecorder, name, expected string) {
	t.Helper()
	cookies := map[string]*http.Cookie{}
	for _, c := range w.Result().Cookies() {
		cookies[c.Name] = c
	}
	c, ok := cookies[name]
	if expected != NO_COOKIE {
		if !ok {
			t.Errorf("expected cookie %q = %q, not set", name, expected)
		} else if c.Value != expected {
			t.Errorf("cookie %q = %q, want %q", name, c.Value, expected)
		}
	} else {
		if ok {
			t.Errorf("expected no cookie %q, got %q", name, c.Value)
		}
	}
}

func TestCreateSession(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		logIn          func(ctx context.Context, userID string, createdAt time.Time) (*domain.User, *domain.Session, error)
		expectedStatus int
		expectedToken  string
		expectedExpiry string
		expectedUser   string
	}{
		{"invalid input", `{"user": "me"}`, func(ctx context.Context, userID string, createdAt time.Time) (*domain.User, *domain.Session, error) {
			return nil, nil, nil
		}, http.StatusBadRequest, NO_COOKIE, NO_COOKIE, NO_COOKIE},
		{"empty input", `{}`, func(ctx context.Context, userID string, createdAt time.Time) (*domain.User, *domain.Session, error) {
			return &domain.User{}, &domain.Session{}, nil
		}, http.StatusBadRequest, NO_COOKIE, NO_COOKIE, NO_COOKIE},
		{"works correct", `{"user_id": "alex"}`, func(ctx context.Context, userID string, createdAt time.Time) (*domain.User, *domain.Session, error) {
			return &domain.User{
					UserId:      "alex",
					DisplayName: "alex",
					CreatedAt:   Date1,
					IsActive:    true,
				}, &domain.Session{
					Token:     "token",
					CreatedAt: Date1,
					ExpiresAt: Date2,
				}, nil
		}, http.StatusOK, "token", Date2.String(), "alex"},
		{"user not found", `{"user_id": "alex"}`, func(ctx context.Context, userID string, createdAt time.Time) (*domain.User, *domain.Session, error) {
			return nil, nil, domain.ErrUserNotFound
		}, http.StatusUnauthorized, NO_COOKIE, NO_COOKIE, NO_COOKIE},
		{"database error", `{"user_id": "alex"}`, func(ctx context.Context, userID string, createdAt time.Time) (*domain.User, *domain.Session, error) {
			return nil, nil, domain.ErrDatabase
		}, http.StatusInternalServerError, NO_COOKIE, NO_COOKIE, NO_COOKIE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := AuthServiceStub{logIn: tt.logIn}
			h := Handler{cfg: &config.Config{Production: true}, authService: stub}

			r := httptest.NewRequest(http.MethodPost, "/auth", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h.CreateSession(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.expectedStatus)
			}

			checkCookie(t, w, domain.SessionCookieName, tt.expectedToken)
			checkCookie(t, w, domain.SessionExpriationCookieName, tt.expectedExpiry)
			checkCookie(t, w, domain.UserIDCookieName, tt.expectedUser)
		})
	}
}

func TestDeleteSession(t *testing.T) {
	tests := []struct {
		name             string
		logOut           func(ctx context.Context, token string) error
		tokenFromRequest func(r *http.Request) (string, bool)
		expectedStatus   int
		expectedToken    string
		expectedExpiry   string
		expectedUser     string
	}{
		{"correct", func(ctx context.Context, token string) error { return nil }, func(r *http.Request) (string, bool) { return "token", true }, http.StatusNoContent, "", "", ""},
		{"no token", func(ctx context.Context, token string) error { return nil }, func(r *http.Request) (string, bool) { return "", false }, http.StatusUnauthorized, NO_COOKIE, NO_COOKIE, NO_COOKIE},
		{"database error", func(ctx context.Context, token string) error { return domain.ErrDatabase }, func(r *http.Request) (string, bool) { return "token", true }, http.StatusInternalServerError, NO_COOKIE, NO_COOKIE, NO_COOKIE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := AuthServiceStub{logOut: tt.logOut, tokenFromRequest: tt.tokenFromRequest}
			h := Handler{cfg: &config.Config{Production: true}, authService: stub}

			r := httptest.NewRequest(http.MethodDelete, "/auth", strings.NewReader(""))
			w := httptest.NewRecorder()
			h.DeleteSession(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.expectedStatus)
			}

			checkCookie(t, w, domain.SessionCookieName, tt.expectedToken)
			checkCookie(t, w, domain.SessionExpriationCookieName, tt.expectedExpiry)
			checkCookie(t, w, domain.UserIDCookieName, tt.expectedUser)
		})
	}
}

func TestGetSession(t *testing.T) {
	tests := []struct {
		name            string
		userFromRequest func(r *http.Request) (*domain.User, bool)
		expectedStatus  int
		expectedUser    *UserResponse
		expectedError   string
	}{
		{"correct", func(r *http.Request) (*domain.User, bool) {
			return &domain.User{
				UserId:      "alex",
				DisplayName: "Alex is awesome",
				CreatedAt:   Date1,
				IsActive:    true,
			}, true
		}, http.StatusOK, &UserResponse{
			UserId:      "alex",
			DisplayName: "Alex is awesome",
			CreatedAt:   Date1,
			IsActive:    true,
		}, ""},
		{"no user", func(r *http.Request) (*domain.User, bool) {
			return nil, false
		}, http.StatusUnauthorized, nil, "user did not exist\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := AuthServiceStub{userFromRequest: tt.userFromRequest}
			h := Handler{authService: stub}

			r := httptest.NewRequest(http.MethodGet, "/auth", strings.NewReader(""))
			w := httptest.NewRecorder()
			h.GetSession(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.expectedStatus)
			}
			var got UserResponse
			if tt.expectedUser == nil {
				if w.Body.String() != tt.expectedError {
					t.Errorf("expected error: %s, got: %s", tt.expectedError, w.Body.String())
				}
			} else {
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if diff := cmp.Diff(*tt.expectedUser, got); diff != "" {
					t.Error(diff)
				}
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		getUser        func(ctx context.Context, userID string) (*domain.User, error)
		expectedStatus int
		expectedUser   *UserResponse
		expectedError  string
	}{
		{"missing user_id", "/user", func(ctx context.Context, userID string) (*domain.User, error) {
			return nil, nil
		}, http.StatusBadRequest, nil, "user_id is required\n"},
		{"user not found", "/user?user_id=alex", func(ctx context.Context, userID string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		}, http.StatusNotFound, nil, "user not found\n"},
		{"database error", "/user?user_id=alex", func(ctx context.Context, userID string) (*domain.User, error) {
			return nil, domain.ErrDatabase
		}, http.StatusInternalServerError, nil, "internal server error\n"},
		{"correct", "/user?user_id=alex", func(ctx context.Context, userID string) (*domain.User, error) {
			return &domain.User{
				UserId:      "alex",
				DisplayName: "Alex",
				CreatedAt:   Date1,
				IsActive:    true,
			}, nil
		}, http.StatusOK, &UserResponse{
			UserId:      "alex",
			DisplayName: "Alex",
			CreatedAt:   Date1,
			IsActive:    true,
		}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := UserServiceStub{getUser: tt.getUser}
			h := Handler{userService: stub}

			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			h.GetUser(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.expectedStatus)
			}
			if tt.expectedUser != nil {
				var got UserResponse
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if diff := cmp.Diff(*tt.expectedUser, got); diff != "" {
					t.Error(diff)
				}
			} else {
				if w.Body.String() != tt.expectedError {
					t.Errorf("expected error: %q, got: %q", tt.expectedError, w.Body.String())
				}
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		createUser     func(ctx context.Context, userID string, displayName string, createdAt time.Time) (*domain.User, error)
		expectedStatus int
		expectedUser   *UserResponse
		expectedError  string
	}{
		{"invalid body", `{bad json}`, func(ctx context.Context, userID string, displayName string, createdAt time.Time) (*domain.User, error) {
			return nil, nil
		}, http.StatusBadRequest, nil, "invalid request body\n"},
		{"unknown fields", `{"user_id": "alex", "display_name": "Alex", "unknown": "field"}`, func(ctx context.Context, userID string, displayName string, createdAt time.Time) (*domain.User, error) {
			return nil, nil
		}, http.StatusBadRequest, nil, "invalid request body\n"},
		{"user already exists", `{"user_id": "alex", "display_name": "Alex"}`, func(ctx context.Context, userID string, displayName string, createdAt time.Time) (*domain.User, error) {
			return nil, domain.ErrUserExists
		}, http.StatusConflict, nil, "user already exists\n"},
		{"database error", `{"user_id": "alex", "display_name": "Alex"}`, func(ctx context.Context, userID string, displayName string, createdAt time.Time) (*domain.User, error) {
			return nil, domain.ErrDatabase
		}, http.StatusInternalServerError, nil, "internal server error\n"},
		{"correct", `{"user_id": "alex", "display_name": "Alex"}`, func(ctx context.Context, userID string, displayName string, createdAt time.Time) (*domain.User, error) {
			return &domain.User{
				UserId:      "alex",
				DisplayName: "Alex",
				CreatedAt:   Date1,
				IsActive:    true,
			}, nil
		}, http.StatusCreated, &UserResponse{
			UserId:      "alex",
			DisplayName: "Alex",
			CreatedAt:   Date1,
			IsActive:    true,
		}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := UserServiceStub{createUser: tt.createUser}
			h := Handler{userService: stub}

			r := httptest.NewRequest(http.MethodPost, "/user", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h.CreateUser(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.expectedStatus)
			}
			if tt.expectedUser != nil {
				var got UserResponse
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if diff := cmp.Diff(*tt.expectedUser, got); diff != "" {
					t.Error(diff)
				}
			} else {
				if w.Body.String() != tt.expectedError {
					t.Errorf("expected error: %q, got: %q", tt.expectedError, w.Body.String())
				}
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name              string
		body              string
		userFromRequest   func(r *http.Request) (*domain.User, bool)
		updateDisplayName func(ctx context.Context, userID string, displayName string) (*domain.User, error)
		expectedStatus    int
		expectedUser      *UserResponse
		expectedError     string
	}{
		{"invalid body", `{bad json}`,
			func(r *http.Request) (*domain.User, bool) { return nil, false },
			func(ctx context.Context, userID string, displayName string) (*domain.User, error) { return nil, nil },
			http.StatusBadRequest, nil, "invalid request body\n"},
		{"unknown fields", `{"display_name": "Alex", "unknown": "field"}`,
			func(r *http.Request) (*domain.User, bool) { return nil, false },
			func(ctx context.Context, userID string, displayName string) (*domain.User, error) { return nil, nil },
			http.StatusBadRequest, nil, "invalid request body\n"},
		{"no user in context", `{"display_name": "Alex"}`,
			func(r *http.Request) (*domain.User, bool) { return nil, false },
			func(ctx context.Context, userID string, displayName string) (*domain.User, error) { return nil, nil },
			http.StatusUnauthorized, nil, "unauthorised\n"},
		{"update error", `{"display_name": "Alex"}`,
			func(r *http.Request) (*domain.User, bool) {
				return &domain.User{UserId: "alex"}, true
			},
			func(ctx context.Context, userID string, displayName string) (*domain.User, error) {
				return nil, domain.ErrUserNotFound
			},
			http.StatusBadRequest, nil, "user did not exist\n"},
		{"correct", `{"display_name": "Alex"}`,
			func(r *http.Request) (*domain.User, bool) {
				return &domain.User{UserId: "alex"}, true
			},
			func(ctx context.Context, userID string, displayName string) (*domain.User, error) {
				return &domain.User{
					UserId:      "alex",
					DisplayName: "Alex",
					CreatedAt:   Date1,
					IsActive:    true,
				}, nil
			},
			http.StatusOK, &UserResponse{
				UserId:      "alex",
				DisplayName: "Alex",
				CreatedAt:   Date1,
				IsActive:    true,
			}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authStub := AuthServiceStub{userFromRequest: tt.userFromRequest}
			userStub := UserServiceStub{updateDisplayName: tt.updateDisplayName}
			h := Handler{authService: authStub, userService: userStub}

			r := httptest.NewRequest(http.MethodPut, "/user", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			h.UpdateUser(w, r)

			if w.Code != tt.expectedStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.expectedStatus)
			}
			if tt.expectedUser != nil {
				var got UserResponse
				if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if diff := cmp.Diff(*tt.expectedUser, got); diff != "" {
					t.Error(diff)
				}
			} else {
				if w.Body.String() != tt.expectedError {
					t.Errorf("expected error: %q, got: %q", tt.expectedError, w.Body.String())
				}
			}
		})
	}
}
