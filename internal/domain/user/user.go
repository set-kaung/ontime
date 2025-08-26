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
	Email            bool      `json:"email"`
	ServicesReceived uint32    `json:"services_received"`
	ServicesProvided uint32    `json:"services_provided"`
	OneTimePaid      bool      `json:"paid"`
	Rating           float32   `json:"rating"`
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
