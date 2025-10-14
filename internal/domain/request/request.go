package request

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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
	ID          int       `json:"event_id"`
	Time        time.Time `json:"event_time"`
	Description string    `json:"event_description"`
	By          string    `json:"by"`
	EventOwner  string    `json:"event_owner"`
}

func (e *Event) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %T to Event", value)
	}
	return json.Unmarshal(bytes, e)
}

func (e Event) Value() (driver.Value, error) {
	return json.Marshal(e)
}

type Events []Event

func (e *Events) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot convert %T to Events", value)
	}
	return json.Unmarshal(bytes, e)
}

func (e Events) Value() (driver.Value, error) {
	return json.Marshal(e)
}

type RequestReport struct {
	ID        int32     `json:"id"`
	UserID    string    `json:"user_id"`
	RequestID int32     `json:"request_id"`
	TicketID  string    `json:"ticket_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `jons:"created_at"`
}
