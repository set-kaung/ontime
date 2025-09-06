package request

import (
	"time"

	"github.com/set-kaung/senior_project_1/internal/domain/listing"
	"github.com/set-kaung/senior_project_1/internal/domain/review"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
)

type RequestType string

const (
	OUTGOING RequestType = "OUTGOING"
	INCOMING RequestType = "INCOMING"
)

type Request struct {
	ID                 int32           `json:"id"`
	Listing            listing.Listing `json:"listing"`
	Requester          user.User       `json:"requester"`
	Provider           user.User       `json:"provider"`
	Activity           string          `json:"activity"`
	StatusDetail       string          `json:"status_detail"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	TokenReward        int32           `json:"token_reward"`
	Type               RequestType     `json:"type"`
	ProviderCompleted  bool            `json:"provider_completed"`
	RequesterCompleted bool            `json:"requester_completed"`
	IsProvider         bool            `json:"is_provider"`
	Review             review.Review   `json:"review"`
	Events             []Event         `json:"events"`
}

func CreateClientServiceRequest(listingID int32, requesterID string) Request {
	return Request{Listing: listing.Listing{ID: listingID}, Requester: user.User{ID: requesterID}}
}

type Event struct {
	ID                int32     `json:"id"`
	Timestamp         time.Time `json:"timestamp"`
	ActionDescription string    `jons:"description"`
}
