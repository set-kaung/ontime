package request

import (
	"time"

	"github.com/set-kaung/senior_project_1/internal/domain/listing"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
)

type Request struct {
	ID           int32
	Listing      listing.Listing `json:"listing"`
	Requester    user.User       `json:"requester"`
	Provider     user.User       `json:"provider"`
	Activity     string          `json:"activity"`
	StatusDetail string          `json:"status_detail"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	TokenReward  int32           `json:"token_reward"`
}

func CreateClientServiceRequest(listingID int32, requesterID string) Request {
	return Request{Listing: listing.Listing{ID: listingID}, Requester: user.User{ID: requesterID}}
}
