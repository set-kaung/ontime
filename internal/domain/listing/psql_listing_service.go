package listing

import (
	"context"
	"log"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/domain/review"
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
			ImageURL: dbListing.ImageUrl.String,
			Status:   dbListing.Status,
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
	repo := repository.New(pls.DB).WithTx(tx)
	createListingParams := repository.InsertListingParams{}
	createListingParams.Title = listing.Title
	createListingParams.Description = listing.Description
	createListingParams.Category = listing.Category
	createListingParams.TokenReward = listing.TokenReward
	createListingParams.PostedBy = listing.Provider.ID
	createListingParams.ImageUrl = pgtype.Text{String: listing.ImageURL, Valid: listing.ImageURL != ""}
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
			ImageURL:    dbListing.ImageUrl.String,
			Status:      dbListing.Status,
		}
	}
	return listings, nil
}

func (pls *PostgresListingService) GetListingByID(ctx context.Context, id int32, userId string) (Listing, error) {
	repo := repository.New(pls.DB)
	dbListing, err := repo.GetListingByID(ctx,
		repository.GetListingByIDParams{
			ID:          id,
			RequesterID: userId,
		})
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
	listing.Provider = user.User{
		ID:       dbListing.Uid,
		FullName: dbListing.FullName,
		Rating:   float32(dbListing.TotalRatings.Int32) / max(1.0, float32(dbListing.RatingCount.Int32))}
	listing.ImageURL = dbListing.ImageUrl.String
	listing.TakenRequestID = -1
	listing.Status = dbListing.Status

	if dbListing.RequestID.Valid {
		listing.TakenRequestID = dbListing.RequestID.Int32
	}

	return listing, nil
}

func (pls *PostgresListingService) DeleteListing(ctx context.Context, id int32, postedBy string) error {
	tx, err := pls.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("CreateLising: failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(pls.DB).WithTx(tx)
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
		log.Printf("Listing Service -> UpdateListing: failed to start transaction: %s\n", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(pls.DB).WithTx(tx)
	rowsAffected, err := repo.UpdateListing(ctx, repository.UpdateListingParams{
		ID:          l.ID,
		PostedBy:    l.Provider.ID,
		Title:       l.Title,
		Description: l.Description,
		Category:    l.Category,
		TokenReward: l.TokenReward,
		ImageUrl:    pgtype.Text{String: l.ImageURL, Valid: l.ImageURL != ""},
	})
	if err != nil {
		log.Printf("listing_service -> UpdateListing: failed to update listing : %v\n", err)
		return -1, internal.ErrInternalServerError
	}
	if rowsAffected == 0 {
		return -1, internal.ErrUnauthorized
	}
	if err = tx.Commit(ctx); err != nil {
		log.Printf("listing_service -> UpdateListing: failed to commit: %v\n", err)
		return -1, internal.ErrInternalServerError
	}
	return l.ID, nil
}

func (pls *PostgresListingService) ReportListing(ctx context.Context, lr ListingReport) error {
	tx, err := pls.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("listing_service -> ReportListing: failed to start transaction: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(pls.DB).WithTx(tx)
	err = repo.InsertReport(ctx, repository.InsertReportParams{
		ListingID:    lr.ListingID,
		ReporterID:   lr.ReporterID,
		ReportReason: pgtype.Text{String: lr.ReportReason, Valid: true},
		AdditionalDetail: pgtype.Text{
			String: lr.AdditionalDetail,
			Valid:  strings.TrimSpace(lr.AdditionalDetail) != "",
		},
	})
	if err != nil {
		log.Printf("listing_service -> ReportListing: failed to insert report: %s\n", err)
		return internal.ErrInternalServerError
	}
	if err = tx.Commit(ctx); err != nil {
		log.Printf("ReportListing: failed to commit: %v\n", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (p *PostgresListingService) GetListingReviews(ctx context.Context, listingID int32) ([]review.Review, error) {
	repo := repository.New(p.DB)
	dbReviews, err := repo.GetListingReviews(ctx, listingID)
	if err != nil {
		log.Printf("GetListingReviews: failed to get listing reviews: %s\n", err)
		return nil, internal.ErrInternalServerError
	}
	reviews := make([]review.Review, 0, len(dbReviews))
	for _, dbr := range dbReviews {
		reviews = append(reviews,
			review.Review{
				ID:               dbr.ID,
				RequestID:        dbr.RequestID,
				ReviewerID:       dbr.ReviewerID,
				ReviewerFullName: dbr.ReviewerFullName,
				RevieweeFullName: dbr.RevieweeFullName,
				RevieweeID:       dbr.RevieweeID,
				Comment:          dbr.Comment.String,
				Rating:           dbr.Rating,
				CreatedAt:        dbr.DateTime,
			},
		)
	}
	return reviews, nil
}
