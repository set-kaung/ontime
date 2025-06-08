package listing

import (
	"time"

	"github.com/set-kaung/senior_project_1/internal/domain/user"
)

type Listing struct {
	ID          int32     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	TokenReward int32     `json:"token_reward"`
	PostedAt    time.Time `json:"posted_at"`
	Category    string    `json:"category"`
	user.User
}
