package request

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/domain/listing"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type PostgresRequestService struct {
	DB *pgxpool.Pool
}

func (prs *PostgresRequestService) CreateServiceRequest(ctx context.Context, r Request) (int32, error) {
	tx, err := prs.DB.Begin(ctx)
	if err != nil {
		log.Println("CreateServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(tx)
	insertServiceRequestParams := repository.InsertPendingServiceRequestParams{
		ListingID:   r.Listing.ID,
		RequesterID: r.Requester.ID,
	}
	rid, err := repo.InsertPendingServiceRequest(ctx, insertServiceRequestParams)

	if err != nil {
		log.Println("CreateServiceRequest: failed to insert to db: ", err)
		return -1, err
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Println("CreateServiceRequest: failed to commit to db: ", err)
		return -1, internal.ErrInternalServerError
	}
	return rid, nil
}

func (prs *PostgresRequestService) GetAllIncomingRequests(ctx context.Context, provider_id string) ([]Request, error) {
	repo := repository.New(prs.DB)
	dbRequests, err := repo.GetAllIncomingServiceRequests(ctx, provider_id)
	if err != nil {
		log.Println("GetAllIncomingRequests: failed to retrieve from db: ", err)
		return nil, internal.ErrInternalServerError
	}
	requests := make([]Request, len(dbRequests))
	for i := range len(dbRequests) {
		dbRequest := dbRequests[i]
		requests[i] = Request{
			ID:           dbRequest.ID,
			Listing:      listing.Listing{ID: dbRequest.ListingID},
			Requester:    user.User{ID: dbRequest.RequesterID},
			Provider:     user.User{ID: dbRequest.ProviderID},
			Activity:     string(dbRequest.Activity),
			StatusDetail: string(dbRequest.StatusDetail),
			DateTime:     dbRequest.DateTime,
		}
	}
	return requests, nil
}

// return unauthorized err if providerID is not equal to the one in DB
// else internal server error
func (prs *PostgresRequestService) AcceptServiceRequest(ctx context.Context, listingID int32, providerID string) (int32, error) {
	repo := repository.New(prs.DB)

	repoRequest, err := repo.GetRequestByID(ctx, listingID)
	if err != nil {
		log.Println("AcceptServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}

	if repoRequest.ProviderID != providerID && repoRequest.Activity != "active" {
		return -1, internal.ErrUnauthorized
	}

	tx, err := prs.DB.Begin(ctx)
	if err != nil {
		log.Println("AcceptServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	acceptServiceParams := repository.InsertServiceRequestParams{
		ListingID:    listingID,
		RequesterID:  repoRequest.RequesterID,
		ProviderID:   providerID,
		StatusDetail: "accepted",
		Activity:     "active",
	}
	repo = repository.New(tx)
	id, err := repo.InsertServiceRequest(ctx, acceptServiceParams)
	if err != nil {
		log.Println("AcceptServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	return id, nil
}

func (prs *PostgresRequestService) DeclineServiceRequest(ctx context.Context, listingID int32, providerID string) (int32, error) {
	repo := repository.New(prs.DB)

	repoRequest, err := repo.GetRequestByID(ctx, listingID)
	if err != nil {
		log.Println("AcceptServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}

	if repoRequest.ProviderID != providerID && repoRequest.Activity != "active" {
		return -1, internal.ErrUnauthorized
	}

	tx, err := prs.DB.Begin(ctx)
	if err != nil {
		log.Println("AcceptServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	acceptServiceParams := repository.InsertServiceRequestParams{
		ListingID:    listingID,
		RequesterID:  repoRequest.RequesterID,
		ProviderID:   providerID,
		StatusDetail: "declined",
		Activity:     "inactive",
	}
	repo = repository.New(tx)
	id, err := repo.InsertServiceRequest(ctx, acceptServiceParams)
	if err != nil {
		log.Println("AcceptServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	return id, nil

}
