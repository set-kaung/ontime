package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/set-kaung/senior_project_1/internal/repository"
	"github.com/set-kaung/senior_project_1/internal/util"
)

type Profile struct {
	Email           string
	Username        string
	Tokens          int32
	Joined_At       time.Time
	Rating          float64
	ProfilePhotoURL string
}

type UserService struct {
	Repo *repository.Queries
}

func (us *UserService) GetUserByID(ctx context.Context, id int32) (*repository.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}
	result, err := us.Repo.GetUserByID(ctx, id)
	return &result, err
}

func (us *UserService) InsertUser(ctx context.Context, email Email, username, password string) error {
	hashedPass, err := util.HashPassword(password)
	if err != nil {
		return err
	}
	insertParams := repository.InsertUserParams{Email: email.Address, Password: hashedPass}
	id, err := us.Repo.InsertUser(ctx, insertParams)
	if err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok {
			if pgerr.Code == "23505" {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	insertPorfileParams := repository.InsertProfileParams{UserID: id, Username: username, Tokens: 0, Rating: 0.0}
	_, err = us.Repo.InsertProfile(ctx, insertPorfileParams)
	return err
}

func (us *UserService) GetUserByEmail(ctx context.Context, email Email) (*repository.User, error) {
	user, err := us.Repo.GetUserByEmail(ctx, email.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email(service): %v", err)
	}
	return &user, nil
}

func (us *UserService) GetUserProfile(ctx context.Context, id int32) (*Profile, error) {
	user, err := us.Repo.GetProfile(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id(service) to view: %v", err)
	}
	prof := &Profile{Email: user.Email, Username: user.Username, Tokens: user.Tokens, Joined_At: user.JoinedAt, Rating: user.Rating, ProfilePhotoURL: user.ProfilePhotoUrl.String}
	return prof, nil
}

func (us *UserService) Exists(ctx context.Context, id int32) (bool, error) {
	return us.Repo.UserExists(ctx, id)
}
