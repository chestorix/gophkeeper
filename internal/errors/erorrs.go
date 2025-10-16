package errors

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrDataNotFound         = errors.New("data not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrCreateUserFailed     = errors.New("create user failed")
	ErrGenerateTokenFailed  = errors.New("generate token failed")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrInvalidToken         = errors.New("invalid token")
	ErrGetSecretDataFailed  = errors.New("get secret data failed")
	ErrSaveSecretDataFailed = errors.New("save secret data failed")
)
