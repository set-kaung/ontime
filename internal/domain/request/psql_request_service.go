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

	repo := repository.New(prs.DB).WithTx(tx)

	insertServiceRequestParams := repository.InsertPendingServiceRequestParams{
		ListingID:   r.Listing.ID,
		RequesterID: r.Requester.ID,
	}

	rid, err := repo.InsertPendingServiceRequest(ctx, insertServiceRequestParams)
	if err != nil {
		log.Println("CreateServiceRequest: failed to insert service request to db: ", err)
		return -1, err
	}

	err = repo.InsertServiceRequestCompletion(ctx, rid)
	if err != nil {
		log.Println("CreateServiceRequest: failed to insert service request completion to db: ", err)
		return -1, err
	}

	rowsAffected, err := repo.DeductTokens(ctx, repository.DeductTokensParams{
		ListingID: r.Listing.ID,
		UserID:    r.Requester.ID,
	})

	if rowsAffected != 1 {
		return -1, internal.ErrInsufficientBalance
	}

	if err != nil {
		log.Println("CreateServiceRequest: failed to deduct tokens: ", err)
		return -1, internal.ErrInternalServerError
	}
	insertPaymentRequestParams := repository.InsertPaymentHoldingParams{
		ServiceRequestID: rid,
		PayerID:          r.Requester.ID,
	}

	_, err = repo.InsertPaymentHolding(ctx, insertPaymentRequestParams)
	if err != nil {
		log.Println("CreateServiceRequest: failed to insert payment holding to db: ", err)
		return -1, err
	}
	if err := tx.Commit(ctx); err != nil {
		log.Println("CreateServiceRequest: failed to commit transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	return rid, nil
}

func (prs *PostgresRequestService) GetRequestByID(ctx context.Context, rid int32) (Request, error) {
	repo := repository.New(prs.DB)
	dbRequest, err := repo.GetRequestByID(ctx, rid)
	if err != nil {
		log.Println("GetRequestByID: failed to retrieve from db: ", err)
		return Request{}, internal.ErrInternalServerError
	}
	r := Request{
		ID: dbRequest.SrID,
		Listing: listing.Listing{
			ID:          dbRequest.SlID,
			Title:       dbRequest.SlTitle,
			Description: dbRequest.SlDescription,
			Category:    dbRequest.SlCategory,
		},
		Requester: user.User{
			ID:       dbRequest.RequesterID,
			FullName: dbRequest.RequesterFullName,
			JoinedAt: dbRequest.RequesterJoinedAt,
		},
		Provider: user.User{
			ID:       dbRequest.SrProviderID,
			FullName: dbRequest.ProviderFullName,
			JoinedAt: dbRequest.ProviderJoinedAt,
		},
		CreatedAt:          dbRequest.SrCreatedAt,
		StatusDetail:       string(dbRequest.SrStatusDetail),
		Activity:           string(dbRequest.SrActivity),
		TokenReward:        dbRequest.SrTokenReward,
		ProviderCompleted:  dbRequest.ProviderCompleted,
		RequesterCompleted: dbRequest.RequesterCompleted,
	}
	return r, nil
}

func (prs *PostgresRequestService) GetUserActiveServiceRequests(ctx context.Context, user_id string) ([]Request, error) {
	repo := repository.New(prs.DB)
	dbRequests, err := repo.GetActiveUserServiceRequests(ctx, user_id)
	if err != nil {
		log.Println("GetAllIncomingRequests: failed to retrieve from db: ", err)
		return nil, internal.ErrInternalServerError
	}
	requests := make([]Request, len(dbRequests))
	var requestType RequestType
	for i := range len(dbRequests) {
		dbRequest := dbRequests[i]
		if dbRequest.ProviderID == user_id {
			requestType = INCOMING
		} else {
			requestType = OUTGOING
		}

		requests[i] = Request{
			ID:           dbRequest.ID,
			Listing:      listing.Listing{ID: dbRequest.ListingID, Title: dbRequest.Title},
			Requester:    user.User{ID: dbRequest.RequesterID, FullName: dbRequest.RequesterName},
			Provider:     user.User{ID: dbRequest.ProviderID, FullName: dbRequest.ProviderName},
			Activity:     string(dbRequest.Activity),
			StatusDetail: string(dbRequest.StatusDetail),
			CreatedAt:    dbRequest.CreatedAt,
			UpdatedAt:    dbRequest.UpdatedAt,
			Type:         requestType,
			TokenReward:  dbRequest.TokenReward,
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

	if repoRequest.ProviderID != providerID || repoRequest.SrActivity != "active" {
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
		StatusDetail: "in_progress",
		Activity:     "active",
	}
	repo = repository.New(prs.DB).WithTx(tx)
	id, err := repo.UpdateServiceRequest(ctx, acceptServiceParams)
	if err != nil {
		log.Println("AcceptServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	if err := tx.Commit(ctx); err != nil {
		log.Println("AcceptServiceRequest: failed to commit transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	return id, nil
}

func (prs *PostgresRequestService) DeclineServiceRequest(ctx context.Context, requestID int32, providerID string) (int32, error) {
	repo := repository.New(prs.DB)

	// make sure the declien came from the provider and that the request is active
	repoRequest, err := repo.GetRequestByID(ctx, requestID)
	if err != nil {
		log.Println("DeclineServiceRequest: failed to get request from db: ", err)
		return -1, internal.ErrInternalServerError
	}
	if repoRequest.ProviderID != providerID || repoRequest.SrActivity != "active" {
		return -1, internal.ErrUnauthorized
	}

	tx, err := prs.DB.Begin(ctx)
	if err != nil {
		log.Println("DeclineServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo = repository.New(prs.DB).WithTx(tx)
	rID, err := repo.UpdateServiceRequest(ctx, repository.UpdateServiceRequestParams{
		ID:           requestID,
		StatusDetail: "declined",
		Activity:     "inactive",
	})
	if err != nil {
		log.Println("DeclineServiceRequest: failed to update service request: ", err)
		return -1, internal.ErrInternalServerError
	}
	paymentHolding, err := repo.GetPaymentHolding(ctx, repository.GetPaymentHoldingParams{
		ServiceRequestID: repoRequest.SrID,
		PayerID:          repoRequest.RequesterID,
	})
	if err != nil {
		log.Println("DeclineServiceRequest: failed to get payment holding: ", err)
		return -1, internal.ErrInternalServerError
	}

	err = repo.AddTokens(ctx, repository.AddTokensParams{
		TokenBalance: paymentHolding.AmountTokens,
		ID:           repoRequest.RequesterID,
	})

	if err != nil {
		log.Println("DeclineServiceRequest: failed to add user tokens: ", err)
		return -1, internal.ErrInternalServerError
	}
	_, err = repo.UpdatePaymentHolding(ctx, repository.UpdatePaymentHoldingParams{
		Status:           "refunded",
		ServiceRequestID: rID,
	})

	if err != nil {
		log.Println("DeclineServiceRequest: failed to update payment holding: ", err)
		return -1, internal.ErrInternalServerError
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("DeclineServiceRequest: failed to commit transaction: ", err)
		return -1, internal.ErrInternalServerError
	}

	return rID, nil

}

func (prs *PostgresRequestService) CompleteServiceRequest(ctx context.Context, requestID int32, userID string) (int32, error) {
	rid := int32(-1)
	tx, err := prs.DB.Begin(ctx)
	repo := repository.New(prs.DB).WithTx(tx)
	if err != nil {
		log.Println("CompleteServiceRequest: failed to create db transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	request, err := repo.GetRequestByID(ctx, requestID)
	if err != nil {
		log.Println("CompleteServiceRequest: failed to get request from db: ", err)
		return -1, internal.ErrInternalServerError
	}
	if request.ProviderID != userID && request.RequesterID != userID {
		return -1, internal.ErrUnauthorized
	}

	requestCompletion, err := repo.GetServiceRequestCompletion(ctx, requestID)
	if err != nil {
		log.Println("CompleteServiceRequest: failed to get requestCompletion from db: ", err)
		return -1, internal.ErrInternalServerError
	}
	if !requestCompletion.IsActive {
		return -1, internal.ErrUnauthorized
	}
	requesterComplete := requestCompletion.RequesterCompleted || (userID == request.RequesterID)
	providerComplete := requestCompletion.ProviderCompleted || (userID == request.ProviderID)
	err = repo.UpdateServiceRequestCompletion(ctx, repository.UpdateServiceRequestCompletionParams{
		RequesterCompleted: requesterComplete,
		ProviderCompleted:  providerComplete,
		IsActive:           !(requesterComplete && providerComplete),
		RequestID:          requestID,
	})
	if err != nil {
		log.Println("CompleteServiceRequest: failed to get requestCompletion from db: ", err)
		return -1, internal.ErrInternalServerError
	}
	requestCompletion, err = repo.GetServiceRequestCompletion(ctx, requestID)
	if err != nil {
		log.Println("CompleteServiceRequest: failed to get requestCompletion from db: ", err)
		return -1, internal.ErrInternalServerError
	}

	if requestCompletion.RequesterCompleted && requestCompletion.ProviderCompleted {
		paymentHolding, err := repo.GetPaymentHolding(ctx, repository.GetPaymentHoldingParams{
			ServiceRequestID: request.SrID,
			PayerID:          request.RequesterID,
		})
		if err != nil {
			log.Println("CompleteServiceRequest: failed to get payment holding: ", err)
			return -1, internal.ErrInternalServerError
		}
		err = repo.AddTokens(ctx, repository.AddTokensParams{
			TokenBalance: paymentHolding.AmountTokens,
			ID:           request.ProviderID,
		})
		if err != nil {
			log.Println("CompleteServiceRequest: failed to get add user tokens: ", err)
			return -1, internal.ErrInternalServerError
		}

		_, err = repo.UpdatePaymentHolding(ctx, repository.UpdatePaymentHoldingParams{
			Status:           "released",
			ServiceRequestID: requestID,
		})

		if err != nil {
			log.Println("CompleteServiceRequest: failed to update payment holding ", err)
			return -1, internal.ErrInternalServerError
		}

		rid, err = repo.UpdateServiceRequest(ctx, repository.UpdateServiceRequestParams{
			StatusDetail: "completed",
			Activity:     "inactive",
			ID:           requestID,
		})

		if err != nil {
			log.Println("CompleteServiceRequest: failed to update payment holding ", err)
			return -1, internal.ErrInternalServerError
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("CompleteServiceRequest: failed to commit transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	return rid, nil
}
