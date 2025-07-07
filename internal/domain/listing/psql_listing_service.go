package listing

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type PostgresListingService struct {
	DB *pgxpool.Pool
}

func (pls *PostgresListingService) GetAllListings(ctx context.Context, postedBy string) ([]Listing, error) {
	repo := repository.New(pls.DB)
	dbListings, err := repo.GetAllListings(ctx, postedBy)
	if err != nil {
		log.Println("psql_listing_service -> GetAllListings: err getting all listings : ", err)
		return nil, err
	}
	listings := make([]Listing, len(dbListings))
	for i := range len(dbListings) {
		dbListing := dbListings[i]

		listings[i] = Listing{
			ID:          dbListing.ID,
			Title:       dbListing.Title,
			Description: dbListing.Description,
			TokenReward: dbListing.TokenReward,
			PostedAt:    dbListing.PostedAt,
			Category:    dbListing.Category,
			Provider: user.User{
				ID:       dbListing.Uid,
				FullName: dbListing.FullName,
			},
		}
	}
	return listings, nil
}

func (pls *PostgresListingService) CreateListing(ctx context.Context, listing Listing) (int32, error) {
	tx, err := pls.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("CreateLising: failed to begin tx: %s\n", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(tx)
	createListingParams := repository.InsertListingParams{}
	createListingParams.Title = listing.Title
	createListingParams.Description = listing.Description
	createListingParams.Category = listing.Category
	createListingParams.TokenReward = listing.TokenReward
	createListingParams.PostedBy = listing.Provider.ID
	id, err := repo.InsertListing(ctx, createListingParams)
	if err != nil {
		log.Printf("ListingService -> CreateListing : error creating listing: %s\n", err)
		return -1, internal.ErrInternalServerError
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("ListingService -> CreateListing : error commiting: %s\n", err)
		return -1, internal.ErrInternalServerError
	}
	return id, nil
}

func (pls *PostgresListingService) GetListingsByUserID(ctx context.Context, postedBy string) ([]Listing, error) {
	repo := repository.New(pls.DB)
	dbListings, err := repo.GetUserListings(ctx, postedBy)
	if err != nil {
		log.Printf("ListingService -> GetUserListing: err getting user listing: %s\n", err)
		return nil, internal.ErrInternalServerError
	}
	listings := make([]Listing, len(dbListings))
	for i := range len(dbListings) {
		dbListing := dbListings[i]

		listings[i] = Listing{
			ID:          dbListing.ID,
			Title:       dbListing.Title,
			Description: dbListing.Description,
			TokenReward: dbListing.TokenReward,
			Category:    dbListing.Category,
			PostedAt:    dbListing.PostedAt,
		}
	}
	return listings, nil
}

func (pls *PostgresListingService) GetListingByID(ctx context.Context, id int32) (Listing, error) {
	repo := repository.New(pls.DB)
	dbListing, err := repo.GetListingByID(ctx, id)
	if err != nil {
		log.Println("psql_listing_service -> GetListingByID: err getting listing by id: ", err)
		return Listing{}, err
	}
	listing := Listing{}
	listing.ID = dbListing.ID
	listing.Title = dbListing.Title
	listing.Description = dbListing.Description
	listing.Category = dbListing.Category
	listing.TokenReward = dbListing.TokenReward
	listing.PostedAt = dbListing.PostedAt
	listing.Provider = user.User{ID: dbListing.Uid, FullName: dbListing.FullName}
	return listing, nil
}

func (pls *PostgresListingService) DeleteListing(ctx context.Context, id int32, postedBy string) error {
	tx, err := pls.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("CreateLising: failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(tx)
	candidateListing := repository.DeleteListingParams{ID: id, PostedBy: postedBy}
	cmdTag, err := repo.DeleteListing(ctx, candidateListing)
	log.Println(cmdTag)
	if err != nil {
		log.Printf("ListingService -> DeleteListing: failed to delete listing: %s\n", err)
		return internal.ErrInternalServerError
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("ListingService -> DeleteListing: failed to delete listing: %s\n", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (pls *PostgresListingService) UpdateListing(ctx context.Context, l Listing) (int32, error) {
	tx, err := pls.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("Listing Service -> Update Listing: failed to delete listinh: %s\n", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(tx)
	rowsAffected, err := repo.UpdateListing(ctx, repository.UpdateListingParams{
		ID:          l.ID,
		PostedBy:    l.Provider.ID,
		Title:       l.Title,
		Description: l.Description,
		Category:    l.Category,
		TokenReward: l.TokenReward,
	})
	if err != nil {
		log.Printf("listing_service -> UpdateListing: err : %v\n", err)
		return -1, internal.ErrInternalServerError
	}
	if rowsAffected == 0 {
		return -1, internal.ErrUnauthorized
	}
	if err = tx.Commit(ctx); err != nil {
		log.Printf("failed to commit: %v\n", err)
		return -1, internal.ErrInternalServerError
	}
	return l.ID, nil
}
