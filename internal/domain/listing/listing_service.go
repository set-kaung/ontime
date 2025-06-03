package listing

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type listingService struct {
	db *pgxpool.Pool
}

func NewListingService(db *pgxpool.Pool) *listingService {
	return &listingService{db: db}
}

func (ls *listingService) GetAllListings(ctx context.Context, postedBy string) ([]repository.ServiceListing, error) {
	repo := repository.New(ls.db)
	return repo.GetAllListings(ctx, postedBy)
}

func (ls *listingService) CreateListing(ctx context.Context, service_listing repository.CreateListingParams) (int32, error) {
	tx, err := ls.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("CreateLising: failed to begin tx: %s\n", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(tx)
	id, err := repo.CreateListing(ctx, service_listing)
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

func (ls *listingService) GetUserListings(ctx context.Context, postedBy string) ([]repository.ServiceListing, error) {
	repo := repository.New(ls.db)
	listings, err := repo.GetUserListings(ctx, postedBy)
	if err != nil {
		log.Printf("ListingService -> GetUserListing: err getting user listing: %s\n", err)
		return nil, internal.ErrInternalServerError
	}
	return listings, nil
}

func (ls *listingService) GetListingByID(ctx context.Context, id int32) (repository.ServiceListing, error) {
	repo := repository.New(ls.db)
	return repo.GetListingByID(ctx, id)
}

func (ls *listingService) DeleteListing(ctx context.Context, id int32, postedBy string) error {
	tx, err := ls.db.BeginTx(ctx, pgx.TxOptions{})
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
