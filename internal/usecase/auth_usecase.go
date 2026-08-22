package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/auth"
)

type AuthUsecase struct {
	userRepo   domain.UserRepository
	hasher     *auth.PasswordHasher
	jwtService *auth.JWTService
}

func NewAuthUsecase(userRepo domain.UserRepository, hasher *auth.PasswordHasher, jwtService *auth.JWTService) *AuthUsecase {
	return &AuthUsecase{
		userRepo:   userRepo,
		hasher:     hasher,
		jwtService: jwtService,
	}
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  domain.User `json:"user"`
}

func (u *AuthUsecase) NeedsSetup() (bool, error) {
	count, err := u.userRepo.Count()
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (u *AuthUsecase) SetupAdmin(username, password string) (*AuthResponse, error) {
	needsSetup, err := u.NeedsSetup()
	if err != nil {
		return nil, err
	}
	if !needsSetup {
		return nil, errors.New("admin user is already configured")
	}

	if len(username) < 3 || len(password) < 6 {
		return nil, errors.New("username must be >= 3 characters and password >= 6 characters")
	}

	hash, err := u.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New().String(),
		Username:     username,
		PasswordHash: hash,
		IsAdmin:      true,
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, err
	}

	token, err := u.jwtService.GenerateToken(user, 30*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (u *AuthUsecase) Login(username, password string) (*AuthResponse, error) {
	user, err := u.userRepo.GetByUsername(username)
	if err != nil {
		return nil, domain.ErrInvalidPassword
	}

	if !u.hasher.Compare(user.PasswordHash, password) {
		return nil, domain.ErrInvalidPassword
	}

	token, err := u.jwtService.GenerateToken(user, 30*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (u *AuthUsecase) GetMe(userID string) (*domain.User, error) {
	return u.userRepo.GetByID(userID)
}
