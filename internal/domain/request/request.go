package request

import (
	"time"

	"github.com/set-kaung/senior_project_1/internal/domain/listing"
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
	Review             Review          `json:"review"`
}

func CreateClientServiceRequest(listingID int32, requesterID string) Request {
	return Request{Listing: listing.Listing{ID: listingID}, Requester: user.User{ID: requesterID}}
}

type Review struct {
	ID         int32  `json:"id"`
	RequestID  int32  `json:"request_id"`
	ReviewerID string `json:"reviewer_id"`
	RevieweeID string `json:"reviewee_id"`
	Comment    string `json:"comment"`
	Rating     int32  `json:"rating"`
}
