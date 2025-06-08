package listing

import "context"

type ListingService interface {
	CreateListing(context.Context, Listing) (int32, error)
	GetAllListings(context.Context, string) ([]Listing, error)
	GetListingByID(context.Context, int32) (Listing, error)
	GetListingsByUserID(context.Context, string) ([]Listing, error)
	DeleteListing(context.Context, int32, string) error
}
