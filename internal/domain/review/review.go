package review

import "time"

type Review struct {
	ID               int32     `json:"id"`
	RequestID        int32     `json:"request_id"`
	ReviewerID       string    `json:"reviewer_id"`
	ReviewerFullName string    `json:"reviewer_full_name"`
	RevieweeFullName string    `json:"reviewee_full_name"`
	RevieweeID       string    `json:"reviewee_id"`
	Comment          string    `json:"comment"`
	Rating           int32     `json:"rating"`
	CreatedAt        time.Time `json:"created_at"`
}
