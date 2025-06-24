package internal

import "errors"

// Specific errors used throughout the code base
// for consistency

var (
	ErrNoRecord            = errors.New("no matching record found")
	ErrDuplicateID         = errors.New("user with ID already exist")
	ErrInternalServerError = errors.New("internal server error")
	ErrInsufficientBalance = errors.New("not enough tokens")
	ErrUnauthorized        = errors.New("unauthorized")
)
