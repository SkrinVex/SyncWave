package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/syncwave/syncwave/internal/infrastructure/auth"
)

type contextKey string

const UserContextKey contextKey = "user_claims"

type AuthMiddleware struct {
	jwtService *auth.JWTService
}

func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ""

		// 1. Check Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. Fallback to query parameter (essential for audio/img streaming)
		if tokenStr == "" {
			tokenStr = r.URL.Query().Get("token")
		}

		if tokenStr == "" {
			http.Error(w, `{"error":"unauthorized","message":"missing authentication token"}`, http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtService.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, `{"error":"unauthorized","message":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserClaims(ctx context.Context) *auth.Claims {
	claims, ok := ctx.Value(UserContextKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}
