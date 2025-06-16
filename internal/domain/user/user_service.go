package user

import (
	"context"
)

type UserService interface {
	GetUserByID(context.Context, string) (User, error)
	InsertUser(context.Context, User) error
	UpdateUser(context.Context, User) error
}
