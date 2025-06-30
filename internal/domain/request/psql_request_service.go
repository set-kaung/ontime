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

// return InsuffcientBalance error
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
		log.Println("CreateServiceRequest: failed to insert service request to db: ", err)
		return -1, err
	}

	result, err := repo.DeductTokens(ctx, repository.DeductTokensParams{
		TokenBalance: r.Listing.TokenReward,
		ID:           r.Requester.ID,
	})
	if result == 0 {
		return -1, internal.ErrInsufficientBalance
	}
	if err != nil {
		log.Println("CreateServiceRequest: failed to deduct tokens: ", err)
		return -1, internal.ErrInternalServerError
	}
	insertPaymentRequestParams := repository.InsertPaymentHoldingParams{
		ServiceRequestID: rid,
		PayerID:          r.Requester.ID,
		AmountTokens:     r.Listing.TokenReward,
	}

	_, err = repo.InsertPaymentHolding(ctx, insertPaymentRequestParams)
	if err != nil {
		log.Println("CreateServiceRequest: failed to insert payment holding to db: ", err)
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
			CreatedAt:    dbRequest.CreatedAt,
			UpdatedAt:    dbRequest.UpdatedAt,
		}
	}
	return requests, nil
}

// return unauthorized err if providerID is not equal to the one in DB
// else internal server error
func (prs *PostgresRequestService) AcceptServiceRequest(ctx context.Context, requestID int32, providerID string) (int32, error) {
	repo := repository.New(prs.DB)

	repoRequest, err := repo.GetRequestByID(ctx, requestID)
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
	acceptServiceParams := repository.UpdateServiceRequestParams{
		ID:           requestID,
		StatusDetail: "accepted",
		Activity:     "active",
	}
	repo = repository.New(tx)
	id, err := repo.UpdateServiceRequest(ctx, acceptServiceParams)
	if err != nil {
		log.Println("AcceptServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	tx.Commit(ctx)
	return id, nil
}

func (prs *PostgresRequestService) DeclineServiceRequest(ctx context.Context, requestID int32, providerID string) (int32, error) {
	repo := repository.New(prs.DB)

	// make sure the declien request came from the provider and that the request is active
	repoRequest, err := repo.GetRequestByID(ctx, requestID)
	if err != nil {
		log.Println("DeclineServiceRequest: failed to get request from db: ", err)
		return -1, internal.ErrInternalServerError
	}
	if repoRequest.ProviderID != providerID && repoRequest.Activity != "active" {
		return -1, internal.ErrUnauthorized
	}

	tx, err := prs.DB.Begin(ctx)
	if err != nil {
		log.Println("DeclineServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo = repository.New(tx)
	id, err := repo.UpdateServiceRequest(ctx, repository.UpdateServiceRequestParams{
		ID:           requestID,
		StatusDetail: "declined",
		Activity:     "inactive",
	})
	if err != nil {
		log.Println("DeclineServiceRequest: failed to update service request: ", err)
		return -1, internal.ErrInternalServerError
	}
	paymenetHolding, err := repo.GetPaymentHolding(ctx, repository.GetPaymentHoldingParams{
		ServiceRequestID: repoRequest.ID,
		PayerID:          repoRequest.RequesterID,
	})
	if err != nil {
		log.Println("DeclineServiceRequest: failed to get payment holding: ", err)
		return -1, internal.ErrInternalServerError
	}

	err = repo.AddTokens(ctx, repository.AddTokensParams{
		TokenBalance: paymenetHolding.AmountTokens,
		ID:           repoRequest.RequesterID,
	})

	if err != nil {
		log.Println("DeclineServiceRequest: failed to add user tokens: ", err)
		return -1, internal.ErrInternalServerError
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("DeclineServiceRequest: failed to commit transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	return id, nil

}

func (prs *PostgresRequestService) CompleteServiceRequest(ctx context.Context, requestID int32) (int32, error) {
	return -1, nil
}
