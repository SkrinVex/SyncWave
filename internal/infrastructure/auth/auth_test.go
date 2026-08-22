package auth_test

import (
	"testing"
	"time"

	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/auth"
)

func TestPasswordHasher(t *testing.T) {
	hasher := auth.NewPasswordHasher()
	password := "mySuperSecret123!"

	hash, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("unexpected error hashing: %v", err)
	}

	if !hasher.Compare(hash, password) {
		t.Fatalf("expected password to match hash")
	}

	if hasher.Compare(hash, "wrongPassword") {
		t.Fatalf("expected wrong password to fail")
	}
}

func TestJWTService(t *testing.T) {
	secret := "test-secret-key-1234567890123456"
	jwtService := auth.NewJWTService(secret)

	user := &domain.User{
		ID:       "u-12345",
		Username: "musicfan",
		IsAdmin:  true,
	}

	token, err := jwtService.GenerateToken(user, 1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != user.ID || claims.Username != user.Username || !claims.IsAdmin {
		t.Fatalf("claims mismatch: %+v", claims)
	}

	// Test invalid token
	_, err = jwtService.ValidateToken("invalid.token.string")
	if err == nil {
		t.Fatalf("expected error for invalid token")
	}
}
