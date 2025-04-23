package repository

import (
	"errors"

	"github.com/set-kaung/senior_project_1/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	GetUserByID(id int) (*domain.User, error)
	GetUserByUsername(username string) (*domain.User, error)
	GetUserByEmail(email string) (*domain.User, error)
	InsertUser(user *domain.User) error

	GetUserProfile(id int) (*domain.User, error)
	Exists(int) (bool, error)
}
