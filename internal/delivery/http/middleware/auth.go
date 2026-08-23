package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/auth"
)

type contextKey string

const (
	UserContextKey contextKey = "user_claims"
	UserIDKey      contextKey = "user_id"
)

type AuthMiddleware struct {
	jwtService *auth.JWTService
	userRepo   domain.UserRepository
}

func NewAuthMiddleware(jwtService *auth.JWTService, userRepo domain.UserRepository) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		userRepo:   userRepo,
	}
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

		// Verify user actually still exists in database (instant revoking for deleted users)
		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil || user == nil {
			http.Error(w, `{"error":"unauthorized","message":"user account has been deleted or deactivated"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "is_admin", user.IsAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAdmin, ok := r.Context().Value("is_admin").(bool)
		if !ok || !isAdmin {
			http.Error(w, `{"error":"forbidden","message":"administrator privileges required"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUserClaims(ctx context.Context) *auth.Claims {
	claims, ok := ctx.Value(UserContextKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}

func GetUserID(ctx context.Context) string {
	if claims := GetUserClaims(ctx); claims != nil {
		return claims.UserID
	}
	if id, ok := ctx.Value("user_id").(string); ok {
		return id
	}
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}
