package domain

import "errors"

var (
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrOrgNotFound        = errors.New("organisation not found")
	ErrVerificationFailed = errors.New("organisation verification failed")
)
