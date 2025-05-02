package user

import "errors"

var (
	ErrNoRecord       = errors.New("no matching record found")
	ErrDuplicateEmail = errors.New("user with email already exist")
)
