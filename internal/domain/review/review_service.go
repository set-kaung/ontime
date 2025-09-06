package review

import "context"

type ReviewService interface {
	InsertReview(context.Context, Review) (int32, error)
	GetReviewByID(id context.Context, reviewID int32) (Review, error)
}
