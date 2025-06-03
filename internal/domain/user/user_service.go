package user

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type userService struct {
	db *pgxpool.Pool
}

func (us *userService) GetUserByID(ctx context.Context, id string) (repository.User, error) {
	repo := repository.New(us.db)
	result, err := repo.GetUserByID(ctx, id)
	return result, err
}

func (us *userService) InsertUser(ctx context.Context, user repository.InsertUserParams) error {
	tx, err := us.db.Begin(ctx)
	if err != nil {
		log.Printf("failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	repo := repository.New(tx)
	_, err = repo.InsertUser(ctx, user)
	if err != nil {
		log.Println(err)
		if pgerr, ok := err.(*pgconn.PgError); ok {
			if pgerr.Code == "23505" {
				return internal.ErrDuplicateID
			}
		}
		return internal.ErrInternalServerError
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("UserService -> InsertUser: error commiting transaction: %s\n", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (us *userService) GetUserProfile(ctx context.Context, id string) (repository.User, error) {
	repo := repository.New(us.db)
	user, err := repo.GetUserByID(ctx, id)
	if err != nil {
		log.Printf("failed to get user by id(service) to view: %s", err)
		return user, internal.ErrInternalServerError
	}
	return user, nil
}
