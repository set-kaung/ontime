package service

import (
	"errors"
	"fmt"

	"github.com/set-kaung/senior_project_1/internal/domain"
	"github.com/set-kaung/senior_project_1/internal/repository"
	"github.com/set-kaung/senior_project_1/internal/util"
)

type UserService struct {
	Repo repository.UserRepository
}

func (us *UserService) GetUserByID(id int) (*domain.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return us.Repo.GetUserByID(id)
}

func (us *UserService) InsertUser(email domain.Email, username, password string) error {
	hashedPass, err := util.HashPassword(password)
	if err != nil {
		return err
	}
	newUser := &domain.User{Email: email, PasswordHash: hashedPass, Profile: domain.Profile{Username: username, Tokens: 0}}
	err = us.Repo.InsertUser(newUser)
	if err != nil && errors.Is(err, repository.ErrDuplicateEmail) {
		return fmt.Errorf("email already in use")
	} else if err != nil {
		return fmt.Errorf("failed to insert user: %s", err.Error())
	}

	return nil
}

func (us *UserService) GetUserByEmail(email domain.Email) (*domain.User, error) {
	user, err := us.Repo.GetUserByEmail(email.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email(service): %v", err)
	}
	return user, nil
}

func (us *UserService) GetUserProfile(id int) (*domain.User, error) {
	user, err := us.Repo.GetUserProfile(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id(service) to view: %v", err)
	}
	return user, nil
}

func (us *UserService) Exists(id int) (bool, error) {
	return us.Repo.Exists(id)
}
