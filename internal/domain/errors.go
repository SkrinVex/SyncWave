package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInternal          = errors.New("internal server error")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid username or password")
	ErrNoAdminExists     = errors.New("no admin user exists, setup required")
	ErrSyncAlreadyActive = errors.New("sync job is already in progress")
)
