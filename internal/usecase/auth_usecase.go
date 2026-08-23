package usecase

import (
	"errors"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/syncwave/syncwave/internal/domain"
	"github.com/syncwave/syncwave/internal/infrastructure/auth"
)

type AuthUsecase struct {
	userRepo     domain.UserRepository
	settingsRepo domain.SettingsRepository
	hasher       *auth.PasswordHasher
	jwtService   *auth.JWTService
}

func NewAuthUsecase(
	userRepo domain.UserRepository,
	settingsRepo domain.SettingsRepository,
	hasher *auth.PasswordHasher,
	jwtService *auth.JWTService,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
		hasher:       hasher,
		jwtService:   jwtService,
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

func (u *AuthUsecase) IsRegistrationAllowed() (bool, error) {
	val, err := u.settingsRepo.Get("allow_registration")
	if err != nil {
		return false, nil
	}
	return val == "1" || val == "true", nil
}

func (u *AuthUsecase) Register(username, password string) (*AuthResponse, error) {
	allowed, err := u.IsRegistrationAllowed()
	if err != nil || !allowed {
		return nil, errors.New("registration is currently disabled by administrator")
	}

	if len(username) < 3 || len(password) < 6 {
		return nil, errors.New("username must be >= 3 characters and password >= 6 characters")
	}

	// Check if username is taken
	if existing, _ := u.userRepo.GetByUsername(username); existing != nil {
		return nil, errors.New("username is already in use")
	}

	hash, err := u.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	// Get default user quota
	defaultQuota := int64(10737418240) // 10 GB
	if quotaStr, err := u.settingsRepo.Get("default_user_quota_bytes"); err == nil && quotaStr != "" {
		if q, err := strconv.ParseInt(quotaStr, 10, 64); err == nil && q > 0 {
			defaultQuota = q
		}
	}

	user := &domain.User{
		ID:                uuid.New().String(),
		Username:          username,
		PasswordHash:      hash,
		IsAdmin:           false,
		StorageQuotaBytes: defaultQuota,
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
