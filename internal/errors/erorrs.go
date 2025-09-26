package errors

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrDataNotFound = errors.New("data not found")
)
