package user

import "time"

type User struct {
	ID               string    `json:"id"`
	FullName         string    `json:"full_name,omitempty"`
	Phone            string    `json:"phone,omitempty"`
	TokenBalance     int32     `json:"token_balance,omitempty"`
	Status           string    `json:"status,omitempty"`
	AddressLine1     string    `json:"address_line_1,omitempty"`
	AddressLine2     string    `json:"address_line_2,omitempty"`
	City             string    `json:"city,omitempty"`
	StateProvince    string    `json:"state_province,omitempty"`
	ZipPostalCode    string    `json:"zip_postal_code,omitempty"`
	Country          string    `json:"country,omitempty"`
	JoinedAt         time.Time `json:"joined_at,omitzero"`
	IsEmailSignedUp  bool      `json:"is_email_signedup"`
	ServicesReceived uint32    `json:"services_received"`
	ServicesProvided uint32    `json:"services_provided"`
	IsPaid           bool      `json:"is_paid"`
	Rating           float32   `json:"rating"`
	AboutMe          string    `json:"about_me"`
}

type Notification struct {
	ID              int32     `json:"id"`
	Message         string    `json:"message"`
	RecipientUserID string    `json:"recipient_user_id"`
	ActionUserID    string    `json:"action_user_id"`
	IsRead          bool      `json:"is_read"`
	EventID         int64     `json:"event_id"`
	CreatedAt       time.Time `json:"created_at"`
	EventType       string    `json:"event_type"`
	EventTargetID   int32     `json:"event_target_id"`
}

type InteractionHistory struct {
	InteractionType string    `json:"interaction_type"`
	Description     string    `json:"description"`
	IsIncoming      bool      `json:"is_incoming"`
	TargetID        int32     `json:"target_id"`
	Amount          int32     `json:"amount"`
	Status          string    `json:"status"`
	Timestamp       time.Time `json:"timestamp"`
}

type PartialListing struct {
	ID          int32   `json:"id"`
	Title       string  `json:"title"`
	AvgRating   float32 `json:"rating"`
	RatingCount int32   `jons:"rating_count"`
	Category    string  `json:"category"`
	TokenReward int32   `json:"token_reward"`
	ImageURL    string  `json:"image_url"`
}

type UserSummary struct {
	User     `json:"user"`
	Listings []PartialListing `json:"active_listings"`
}
