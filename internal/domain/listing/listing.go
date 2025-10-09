package listing

import (
	"time"

	"github.com/set-kaung/senior_project_1/internal/domain/user"
)

type Listing struct {
	ID              int32         `json:"id"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	TokenReward     int32         `json:"token_reward"`
	PostedAt        time.Time     `json:"posted_at"`
	Category        string        `json:"category"`
	ImageURL        string        `json:"image_url"`
	Provider        user.User     `json:"provider"`
	HasAlreadyTaken bool          `json:"has_already_taken"`
	TakenRequestID  int32         `json:"taken_request_id"`
	Status          string        `json:"status"`
	SessionDuration time.Duration `json:"session_duration"`
	ContactMethod   string        `json:"contact_method"`
	AvgRating       float32       `json:"avg_rating"`
}

type ListingReport struct {
	ID               int32     `json:"id"`
	ListingID        int32     `json:"listing_id"`
	ReporterID       string    `json:"reporter_id"`
	Datetime         time.Time `json:"datetime"`
	ReportReason     string    `json:"report_reason"`
	Status           string    `json:"status"`
	AdditionalDetail string    `json:"additional_detail"`
}
