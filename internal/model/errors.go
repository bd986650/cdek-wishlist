package model

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrForbidden         = errors.New("access denied")
	ErrAlreadyReserved   = errors.New("item is already reserved")
)
