package user

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type PostgresUserService struct {
	DB *pgxpool.Pool
}

func (pus *PostgresUserService) GetUserByID(ctx context.Context, id string) (User, error) {
	repo := repository.New(pus.DB)
	repoUser, err := repo.GetUserByID(ctx, id)

	user := User{}
	user.ID = repoUser.ID
	user.FullName = repoUser.FullName
	user.TokenBalance = repoUser.TokenBalance
	user.Status = string(repoUser.Status)
	user.AddressLine1 = repoUser.AddressLine1
	user.AddressLine2 = repoUser.AddressLine2
	user.City = repoUser.City
	user.StateProvince = repoUser.StateProvince
	user.ZipPostalCode = repoUser.ZipPostalCode
	user.Country = repoUser.Country
	user.JoinedAt = repoUser.JoinedAt
	user.Email = repoUser.Email
	user.ServicesProvided = uint32(repoUser.ServicesProvided)
	user.ServicesReceived = uint32(repoUser.ServicesReceived)

	return user, err
}

func (pus *PostgresUserService) InsertUser(ctx context.Context, user User) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Printf("failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	insertUserParams := repository.InsertUserParams{
		ID:            user.ID,
		FullName:      user.FullName,
		Phone:         user.Phone,
		TokenBalance:  user.TokenBalance,
		AddressLine1:  user.AddressLine1,
		AddressLine2:  user.AddressLine2,
		City:          user.City,
		StateProvince: user.StateProvince,
		ZipPostalCode: user.ZipPostalCode,
		Country:       user.Country,
	}

	insertUserParams.Status = repository.AccountStatusActive

	repo := repository.New(pus.DB).WithTx(tx)
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
	tokenReward := os.Getenv("ONETIME_PAYMENT_TOKENS")
	tokens, _ := strconv.Atoi(tokenReward)

	err = repo.AddTokens(ctx, repository.AddTokensParams{
		TokenBalance: int32(tokens),
		ID:           user.ID,
	})
	if err != nil {
		log.Printf("UserService -> InsertUser: error adding user tokens: %s\n", err)
		return internal.ErrInternalServerError
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("UserService -> InsertUser: error commiting transaction: %s\n", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (pus *PostgresUserService) UpdateUser(ctx context.Context, user User) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Printf("failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	updateUserParams := repository.UpdateUserParams{
		FullName:      user.FullName,
		Phone:         user.Phone,
		AddressLine1:  user.AddressLine1,
		AddressLine2:  user.AddressLine2,
		City:          user.City,
		StateProvince: user.StateProvince,
		Country:       user.Country,
		ZipPostalCode: user.ZipPostalCode,
		ID:            user.ID,
	}
	repo := repository.New(pus.DB).WithTx(tx)
	_, err = repo.UpdateUser(ctx, updateUserParams)
	if err != nil {
		log.Printf("UserService -> UpdateUser: error: %s\n", err)
		return internal.ErrInternalServerError
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("UserService -> InsertUser: error commiting transaction: %s\n", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (pus *PostgresUserService) DeleteUser(ctx context.Context, id string) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Printf("failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Printf("failed to rollback tx: %s\n", err)
		}
	}()

	repo := repository.New(pus.DB).WithTx(tx)
	_, err = repo.DeleteUser(ctx, id)
	if err != nil {
		log.Printf("UserService -> DeleteUser: error: %s\n", err)
		return internal.ErrInternalServerError
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("UserService -> DeleteUser: error commiting transaction: %s\n", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (pus *PostgresUserService) InsertAdsHistory(ctx context.Context, userID string) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Printf("failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Printf("failed to rollback tx: %s\n", err)
		}
	}()

	repo := repository.New(pus.DB).WithTx(tx)
	_, err = repo.InsertAdsHistory(ctx, userID)
	if err != nil {
		log.Printf("failed to insert ads history: %s\n", err)
		return internal.ErrInternalServerError
	}
	err = repo.AddTokens(ctx, repository.AddTokensParams{
		TokenBalance: 1,
		ID:           userID,
	})
	if err != nil {
		log.Printf("failed to add token balance for ad watching: %s\n", err)
		return internal.ErrInternalServerError
	}
	if err = tx.Commit(ctx); err != nil {
		log.Printf("failed to commit ads history: %s\n", err)
		return internal.ErrInternalServerError
	}

	return nil
}

func (pus *PostgresUserService) GetAdsHistory(ctx context.Context, userID string) (int64, error) {
	repo := repository.New(pus.DB)
	count, err := repo.GetAdsWatched(ctx, userID)
	if err != nil {
		log.Printf("failed to get ads history: %s\n", err)
		return -1, internal.ErrInternalServerError
	}
	return count, nil
}
