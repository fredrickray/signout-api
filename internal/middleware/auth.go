package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	authsvc "github.com/signout/signout-api/internal/service/auth"
	"github.com/signout/signout-api/pkg/response"
)

type ctxKey string

const UserIDKey ctxKey = "user_id"
const UserEmailKey ctxKey = "user_email"

func Authenticate(tokens *authsvc.TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(strings.ToLower(header), "bearer ") {
				response.Fail(w, http.StatusUnauthorized, "unauthorized", "missing bearer token")
				return
			}
			raw := strings.TrimSpace(header[7:])
			claims, err := tokens.ParseAccessToken(raw)
			if err != nil {
				response.Fail(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
				return
			}
			userID, err := uuid.Parse(claims.Subject)
			if err != nil {
				response.Fail(w, http.StatusUnauthorized, "unauthorized", "invalid token subject")
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				response.Fail(w, http.StatusInternalServerError, "internal_error", "unexpected server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
