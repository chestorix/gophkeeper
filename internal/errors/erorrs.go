package errors

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrDataNotFound        = errors.New("data not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrCreateUserFailed    = errors.New("create user failed")
	ErrGenerateTokenFailed = errors.New("generate token failed")
)
