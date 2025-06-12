package request

import "context"

type RequestService interface {
	CreateServiceRequest(context.Context, Request) (int32, error)
	GetAllIncomingRequests(context.Context, string) ([]Request, error)
}
