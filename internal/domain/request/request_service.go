package request

import "context"

type RequestService interface {
	CreateServiceRequest(context.Context, Request) (int32, error)
	GetAllIncomingRequests(context.Context, string) ([]Request, error)
	GetRequestByID(context.Context, int32) (Request, error)
	AcceptServiceRequest(context.Context, int32, string) (int32, error)
	DeclineServiceRequest(context.Context, int32, string) (int32, error)
	CompleteServiceRequest(context.Context, int32, string) (int32, error)
}
