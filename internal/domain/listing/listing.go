package listing

import (
	"time"

	"github.com/set-kaung/senior_project_1/internal/domain/user"
	"github.com/set-kaung/senior_project_1/internal/repository"
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
	Warning         Warning       `json:"warning,omitzero"`
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

type Warning struct {
	ID        int32                      `json:"id"`
	UserID    string                     `json:"user_id"`
	Severity  repository.WarningSeverity `json:"severity"`
	CreatedAt time.Time                  `json:"created_at"`
	Reason    string                     `json:"reason"`
	ListingID int32                      `json:"listing_id"`
}
