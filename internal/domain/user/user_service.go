package user

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type UserService interface {
	GetUserByID(context.Context, string) (User, error)
	InsertUser(context.Context, User) error
}

type PostgresUserService struct {
	DB *pgxpool.Pool
}

func (us *PostgresUserService) GetUserByID(ctx context.Context, id string) (User, error) {
	repo := repository.New(us.DB)
	repoUser, err := repo.GetUserByID(ctx, id)

	user := User{}
	user.ID = repoUser.ID
	user.FirstName = repoUser.FirstName
	user.LastName = repoUser.LastName
	user.Phone = repoUser.Phone
	user.TokenBalance = repoUser.TokenBalance
	user.Status = string(repoUser.Status)
	user.AddressLine1 = repoUser.AddressLine1
	user.AddressLine2 = repoUser.AddressLine2
	user.City = repoUser.City
	user.StateProvince = repoUser.StateProvince
	user.ZipPostalCode = repoUser.ZipPostalCode
	user.Country = repoUser.Country
	user.JoinedAt = repoUser.JoinedAt

	return user, err
}

func (us *PostgresUserService) InsertUser(ctx context.Context, user User) error {
	tx, err := us.DB.Begin(ctx)
	if err != nil {
		log.Printf("failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	insertUserParams := repository.InsertUserParams{
		ID:            user.ID,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Phone:         user.Phone,
		TokenBalance:  user.TokenBalance,
		AddressLine1:  user.AddressLine1,
		AddressLine2:  user.AddressLine2,
		City:          user.City,
		StateProvince: user.StateProvince,
		ZipPostalCode: user.ZipPostalCode,
		Country:       user.Country,
	}

	switch user.Status {
	case "active":
		insertUserParams.Status = repository.AccountStatusActive
	case "suspended":
		insertUserParams.Status = repository.AccountStatusSuspended
	case "banned":
		insertUserParams.Status = repository.AccountStatusBanned
	default:
		return errors.New("invalid account status")
	}

	repo := repository.New(tx)
	_, err = repo.InsertUser(ctx, insertUserParams)
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

// func (us *PostgresUserService) GetUserProfile(ctx context.Context, id string) (User, error) {
// 	repo := repository.New(us.db)
// 	repoUser, err := repo.GetUserByID(ctx, id)
// 	user := User{}

// 	user.ID = repoUser.ID
// 	user.FirstName = repoUser.FirstName
// 	user.LastName = repoUser.LastName
// 	user.Phone = repoUser.Phone
// 	user.TokenBalance = repoUser.TokenBalance
// 	user.Status = string(repoUser.Status)
// 	user.AddressLine1 = repoUser.AddressLine1
// 	user.AddressLine2 = repoUser.AddressLine2
// 	user.City = repoUser.City
// 	user.StateProvince = repoUser.StateProvince
// 	user.ZipPostalCode = repoUser.ZipPostalCode
// 	user.Country = repoUser.Country
// 	user.JoinedAt = repoUser.JoinedAt

// 	if err != nil {
// 		log.Printf("failed to get user by id(service) to view: %s", err)
// 		return user, internal.ErrInternalServerError
// 	}
// 	return user, nil
// }
