package user

import (
	"errors"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	GetUserByID(id int) (*User, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByEmail(email string) (*User, error)
	InsertUser(user *User) error

	GetUserProfile(id int) (*User, error)
	Exists(int) (bool, error)
}
