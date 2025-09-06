package listing

import (
	"context"

	"github.com/set-kaung/senior_project_1/internal/domain/review"
)

type ListingService interface {
	CreateListing(context.Context, Listing) (int32, error)
	GetAllListings(context.Context, string) ([]Listing, error)
	GetListingByID(context.Context, int32, string) (Listing, error)
	GetListingsByUserID(context.Context, string) ([]Listing, error)
	UpdateListing(context.Context, Listing) (int32, error)
	DeleteListing(context.Context, int32, string) error
	ReportListing(ctx context.Context, lr ListingReport) error
	GetListingReviews(ctx context.Context, listingID int32) ([]review.Review, error)
}
