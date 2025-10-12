package request

import (
	"context"

	"github.com/set-kaung/senior_project_1/internal/domain/review"
)

type RequestService interface {
	CreateServiceRequest(context.Context, Request) (int32, error)
	GetUserActiveServiceRequests(context.Context, string) ([]Request, error)
	GetRequestByID(context.Context, int32) (Request, error)
	AcceptServiceRequest(context.Context, int32, string) (int32, error)
	DeclineServiceRequest(context.Context, int32, string) (int32, error)
	CancelServiceRequest(ctx context.Context, requestID int32, userID string) error
	CompleteServiceRequest(context.Context, int32, string) (int32, error)
	CreateRequestReport(ctx context.Context, requestID int32, userID string) (string, error)
	GetRequestReport(ctx context.Context, requestID int32, reporterID string) (RequestReport, error)
	GetRequestReview(ctx context.Context, requestID int32) (review.Review, error)
	UpdateExpiredRequests(ctx context.Context) error
}
