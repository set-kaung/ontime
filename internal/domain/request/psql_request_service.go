package request

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/domain"
	"github.com/set-kaung/senior_project_1/internal/domain/listing"
	"github.com/set-kaung/senior_project_1/internal/domain/review"
	"github.com/set-kaung/senior_project_1/internal/repository"
	"github.com/set-kaung/senior_project_1/internal/util"

	"github.com/jackc/pgx/v5"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
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

	paymentID, err := repo.InsertPaymentHolding(ctx, insertPaymentRequestParams)
	if err != nil {
		log.Println("CreateServiceRequest: failed to insert payment holding to db: ", err)
		return -1, err
	}

	request, err := repo.GetRequestByID(ctx, rid)
	if err != nil {
		log.Println("CreateServiceRequest: failed to get request: ", err)
		return -1, internal.ErrInternalServerError
	}

	eID, err := repo.InsertEvent(ctx, repository.InsertEventParams{
		TargetID:    request.SrID,
		Type:        domain.REQUEST_EVENT,
		Description: domain.INITIATE_REQUEST,
	})

	if err != nil {
		log.Println("CreateServiceRequest: failed to insert notification event: ", err)
		return -1, internal.ErrInternalServerError
	}

	err = repo.InsertTransaction(ctx, repository.InsertTransactionParams{
		UserID:    r.Requester.ID,
		Type:      domain.DEDUCTION_TRANS,
		PaymentID: paymentID,
	})
	if err != nil {
		log.Println("CreateServiceRequest: failed to insert transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	_, err = repo.InsertNotification(ctx, repository.InsertNotificationParams{
		Message:         fmt.Sprintf("%s has requested your service \"%s\"", request.RequesterFullName, request.SlTitle),
		RecipientUserID: request.ProviderID,
		ActionUserID:    pgtype.Text{String: request.RequesterID, Valid: true},
		EventID:         eID,
	})

	if err != nil {
		log.Println("CreateServiceRequest: failed to insert notification: ", err)
		return -1, internal.ErrInternalServerError
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("CreateServiceRequest: failed to commit transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	err = internal.PusherClient.Trigger(fmt.Sprintf("user-%s", request.ProviderID), "new-notification", nil)
	if err != nil {
		log.Println("CreateServiceRequest: failed to send notification: ", err)
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

	var events []Event
	if len(dbRequest.Events) > 0 {

		var rawEvents []Event
		if err := json.Unmarshal(dbRequest.Events, &rawEvents); err != nil {
			log.Println("GetRequestByID: failed to unmarshal events JSON: ", err)
			events = []Event{}
		} else {
			events = make([]Event, len(rawEvents))
			for i, rawEvent := range rawEvents {
				switch rawEvent.Description {
				case domain.INITIATE_REQUEST:
					rawEvent.By = "requester"
				case domain.ACCEPT_REQUEST:
					rawEvent.By = "provider"
				case domain.CONFIRM_COMPLETION:
					if rawEvent.EventOwner == dbRequest.RequesterID {
						rawEvent.By = "requester"
					} else {
						rawEvent.By = "provider"
					}
				case domain.DECLINE_REQUEST:
					rawEvent.By = "provider"
				}
				events[i] = rawEvent
			}
		}
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
			ID:       dbRequest.ProviderID,
			FullName: dbRequest.ProviderFullName,
			JoinedAt: dbRequest.ProviderJoinedAt,
		},
		CreatedAt:          dbRequest.SrCreatedAt,
		StatusDetail:       string(dbRequest.SrStatusDetail),
		Activity:           string(dbRequest.SrActivity),
		TokenReward:        dbRequest.SrTokenReward,
		ProviderCompleted:  dbRequest.ProviderCompleted,
		RequesterCompleted: dbRequest.RequesterCompleted,
		Events:             events,
	}

	return r, nil
}

func (prs *PostgresRequestService) GetUserActiveServiceRequests(ctx context.Context, userID string) ([]Request, error) {
	repo := repository.New(prs.DB)
	dbRequests, err := repo.GetActiveUserServiceRequests(ctx, userID)
	if err != nil {
		log.Println("GetAllIncomingRequests: failed to retrieve from db: ", err)
		return nil, internal.ErrInternalServerError
	}
	requests := make([]Request, len(dbRequests))
	var requestType RequestType
	for i := range len(dbRequests) {
		dbRequest := dbRequests[i]
		if dbRequest.ProviderID == userID {
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
			IsProvider:   dbRequest.ProviderID == userID,
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

	eID, err := repo.InsertEvent(ctx, repository.InsertEventParams{
		TargetID:    repoRequest.SrID,
		Type:        domain.REQUEST_EVENT,
		Description: domain.ACCEPT_REQUEST,
	})

	if err != nil {
		log.Println("Accept Service Request: failed to insert notification event: ", err)
		return -1, internal.ErrInternalServerError

	}

	_, err = repo.InsertNotification(ctx, repository.InsertNotificationParams{
		Message:         fmt.Sprintf("%s has accepted your request for \"%s\"", repoRequest.ProviderFullName, repoRequest.SlTitle),
		RecipientUserID: repoRequest.RequesterID,
		ActionUserID:    pgtype.Text{String: repoRequest.ProviderID, Valid: true},
		EventID:         eID,
	})

	if err != nil {
		log.Println("Accept Service Request: failed to insert notification: ", err)
		return -1, internal.ErrInternalServerError

	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("AcceptServiceRequest: failed to commit transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	err = internal.PusherClient.Trigger(fmt.Sprintf("user-%s", repoRequest.RequesterID), "new-notification", nil)
	if err != nil {
		log.Println("AcceptServiceRequest: failed to send notification: ", err)
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

	_, err = repo.AddTokens(ctx, repository.AddTokensParams{
		TokenBalance: paymentHolding.AmountTokens,
		ID:           repoRequest.RequesterID,
	})
	if err != nil {
		log.Println("DeclineServiceRequest: failed to add user tokens: ", err)
		return -1, internal.ErrInternalServerError
	}

	err = repo.InsertTransaction(ctx, repository.InsertTransactionParams{
		UserID:    repoRequest.RequesterID,
		Type:      domain.ADDITION_TRANS,
		PaymentID: paymentHolding.ID,
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

	eventID, err := repo.InsertEvent(ctx, repository.InsertEventParams{
		TargetID:    repoRequest.SrID,
		Type:        domain.REQUEST_EVENT,
		Description: domain.DECLINE_REQUEST,
	})

	if err != nil {
		log.Println("DeclineServiceRequest: failed to insert event: ", err)
		return -1, internal.ErrInternalServerError
	}

	_, err = repo.InsertNotification(ctx, repository.InsertNotificationParams{
		Message:         fmt.Sprintf("%s has declined your service request.", repoRequest.ProviderFullName),
		RecipientUserID: repoRequest.RequesterID,
		ActionUserID:    pgtype.Text{String: repoRequest.ProviderID, Valid: true},
		EventID:         eventID,
	})

	if err != nil {
		log.Println("DeclineServiceRequest: failed to insert notification: ", err)
		return -1, internal.ErrInternalServerError
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("DeclineServiceRequest: failed to commit transaction: ", err)
		return -1, internal.ErrInternalServerError
	}

	err = internal.PusherClient.Trigger(fmt.Sprintf("user-%s", repoRequest.RequesterID), "new-notification", nil)
	if err != nil {
		log.Println("DeclineServiceRequest: failed to send notification: ", err)
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
		_, err = repo.AddTokens(ctx, repository.AddTokensParams{
			TokenBalance: paymentHolding.AmountTokens,
			ID:           request.ProviderID,
		})
		if err != nil {
			log.Println("CompleteServiceRequest: failed to add user tokens: ", err)
			return -1, internal.ErrInternalServerError
		}
		err = repo.InsertTransaction(ctx, repository.InsertTransactionParams{
			UserID:    request.ProviderID,
			Type:      domain.ADDITION_TRANS,
			PaymentID: paymentHolding.ID,
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

	eventID, err := repo.InsertEvent(ctx, repository.InsertEventParams{
		TargetID:    request.SrID,
		Type:        domain.REQUEST_EVENT,
		Description: domain.CONFIRM_COMPLETION,
	})
	if err != nil {
		log.Println("CompleteServiceRequest: failed to insert event: ", err)
		return -1, internal.ErrInternalServerError
	}
	var notificationMessage string
	var actionUserID string
	var recipientID string
	if request.RequesterFullName == "" || request.ProviderFullName == "" {
		return -1, internal.ErrInternalServerError
	}
	if userID == request.RequesterID {
		actionUserID = request.RequesterID
		recipientID = request.ProviderID
		notificationMessage = fmt.Sprintf("%s has confirmed completion.", request.RequesterFullName)
	} else {
		actionUserID = request.ProviderID
		recipientID = request.RequesterID
		notificationMessage = fmt.Sprintf("%s has confirmed completion.", request.ProviderFullName)
	}
	_, err = repo.InsertNotification(ctx, repository.InsertNotificationParams{
		Message:         notificationMessage,
		EventID:         eventID,
		ActionUserID:    pgtype.Text{String: actionUserID, Valid: true},
		RecipientUserID: recipientID,
	})

	if err != nil {
		log.Println("CompleteServiceRequest: failed to insert notification: ", err)
		return -1, internal.ErrInternalServerError
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("CompleteServiceRequest: failed to commit transaction: ", err)
		return -1, internal.ErrInternalServerError
	}

	err = internal.PusherClient.Trigger(fmt.Sprintf("user-%s", recipientID), "new-notification", nil)
	if err != nil {
		log.Println("CompleteServiceRequest: failed to send notification: ", err)
	}
	return rid, nil
}

func (prs *PostgresRequestService) CreateRequestReport(ctx context.Context, requestID int32, userID string) (string, error) {
	tx, err := prs.DB.Begin(ctx)
	if err != nil {
		log.Println("InsertRequestReport: failed to create db transaction: ", err)
		return "", internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	repo := repository.New(prs.DB).WithTx(tx)
	dbReport, err := repo.InsertRequestReport(ctx, repository.InsertRequestReportParams{
		ReporterID: userID,
		RequestID:  requestID,
	})
	if err != nil {
		log.Println("InsertRequestReport: failed to insert request report: ", err)
		return "", internal.ErrInternalServerError
	}
	ticketID, err := util.GenerateTicket(int64(dbReport.ID), dbReport.CreatedAt)
	if err != nil {
		log.Println("InsertRequestReport: failed to generate ticket: ", err)
		return "", internal.ErrInternalServerError
	}
	dbTicketID, err := repo.UpdateRequestReportWithTicketID(ctx, repository.UpdateRequestReportWithTicketIDParams{
		TicketID: ticketID,
		ID:       dbReport.ID,
	})
	if err != nil {
		log.Printf("CreateRequestReport: failed to update request report ticket id: %s\n", err)
		return "", internal.ErrInternalServerError
	}
	if ticketID != dbTicketID {
		log.Printf("InsertRequestReport: should not happened: db ticket: %s, server ticket: %s", dbTicketID, ticketID)
		return "", internal.ErrInternalServerError
	}
	if err := tx.Commit(ctx); err != nil {
		log.Println("InsertRequestReport: failed to commit transaction: ", err)
		return "", internal.ErrInternalServerError
	}
	return ticketID, nil
}
func (prs *PostgresRequestService) GetRequestReview(ctx context.Context, requestID int32) (review.Review, error) {
	repo := repository.New(prs.DB)
	dbReview, err := repo.GetReviewByRequestID(ctx, requestID)
	r := review.Review{}
	if err != nil {
		log.Printf("GetRequestReview: failed to get review by request ID:%s\n", err)
		return r, internal.ErrInternalServerError
	}
	r.ID = dbReview.ID
	r.Rating = dbReview.Rating
	r.RequestID = dbReview.RequestID
	r.RevieweeID = dbReview.RevieweeID
	r.ReviewerID = dbReview.ReviewerID
	r.RevieweeFullName = dbReview.RevieweeFullName
	r.ReviewerFullName = dbReview.ReviewerFullName
	r.Comment = dbReview.Comment.String
	r.CreatedAt = dbReview.DateTime

	return r, nil
}

func (prs *PostgresRequestService) GetRequestReport(ctx context.Context, requestID int32, reporterID string) (RequestReport, error) {
	repo := repository.New(prs.DB)
	dbReport, err := repo.GetRequestReport(ctx, repository.GetRequestReportParams{
		RequestID:  requestID,
		ReporterID: reporterID,
	})
	report := RequestReport{}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return report, internal.ErrNoRecord
		}
		log.Printf("GetRequestReport: failed to get report: %s\n", err)
		return report, internal.ErrInternalServerError
	}
	report.ID = dbReport.ID
	report.RequestID = dbReport.RequestID
	report.TicketID = dbReport.TicketID
	report.UserID = dbReport.ReporterID
	report.Status = dbReport.Status
	return report, nil
}

func (prs *PostgresRequestService) UpdateExpiredRequests(ctx context.Context) error {
	tx, err := prs.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf(" UpdateExpiredRequests: failed to start transaction: %s\n", err)
		return err
	}
	defer tx.Rollback(ctx)
	repo := repository.New(prs.DB).WithTx(tx)
	requestersUpdated, err := repo.UpdateExpiredRequest(ctx)
	if err != nil {
		log.Printf(" UpdateExpiredRequests: failed to update expired request: %v\n", err)
		return err
	}
	for _, row := range requestersUpdated {
		_, err = repo.AddTokens(ctx, repository.AddTokensParams{
			TokenBalance: row.TokenReward,
			ID:           row.RequesterID,
		})
		if err != nil {
			log.Printf(" UpdateExpiredRequests: failed to refund user tokens: %v\n", err)
			return err
		}

		_, err = repo.UpdatePaymentHolding(ctx, repository.UpdatePaymentHoldingParams{
			Status:           repository.PaymentStatusRefunded,
			ServiceRequestID: row.ID,
		})
		if err != nil {
			log.Printf(" UpdateExpiredRequests: failed to update payment status: %v\n", err)
			return err
		}
		eventID, err := repo.InsertEvent(ctx, repository.InsertEventParams{
			TargetID:    row.ID,
			Type:        domain.SYSTEM_EVENT,
			Description: domain.REQUEST_EXPIRED,
		})
		if err != nil {
			log.Printf(" UpdateExpiredRequests: failed to insert events: %v\n", err)
			return err
		}
		_, err = repo.InsertNotification(ctx, repository.InsertNotificationParams{
			Message:         fmt.Sprintf("Your request for \"%s\" has expired. Your tokens have been refunded.", row.ListingTitle),
			RecipientUserID: row.RequesterID,
			ActionUserID:    pgtype.Text{Valid: false},
			EventID:         eventID,
		})
		if err != nil {
			log.Printf(" UpdateExpiredRequests: failed to insert notifications: %v\n", err)
			return err
		}
		_, err = repo.InsertNotification(ctx, repository.InsertNotificationParams{
			Message:         fmt.Sprintf("Request from %s has expired for your service \"%s\".", row.RequesterFullName, row.ListingTitle),
			RecipientUserID: row.ProviderID,
			ActionUserID:    pgtype.Text{Valid: false},
			EventID:         eventID,
		})
		if err != nil {
			log.Printf(" UpdateExpiredRequests: failed to insert notifications: %v\n", err)
			return err
		}
		err = internal.PusherClient.Trigger(fmt.Sprintf("user-%s", row.RequesterID), "new-notification", nil)
		if err != nil {
			log.Printf("CancelServiceRequest: failed to push notification: %s\n", err)
		}
		err = internal.PusherClient.Trigger(fmt.Sprintf("user-%s", row.ProviderID), "new-notification", nil)
		if err != nil {
			log.Printf("CancelServiceRequest: failed to push notification: %s\n", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Println("UpdateExpiredRequests: failed to commit transaction: ", err)
		return err
	}
	return nil
}

func (prs *PostgresRequestService) CancelServiceRequest(ctx context.Context, requestID int32, userID string) error {
	tx, err := prs.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Printf("CancelServiceRequest: failed to begin db transaction: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(tx)
	repoRequest, err := repo.GetRequestByID(ctx, requestID)
	if err != nil {
		log.Printf("CancelServiceRequest: failed to get request by ID:%s\n", err)
		return internal.ErrInternalServerError
	}
	if repoRequest.RequesterID != userID || repoRequest.SrActivity != repository.ServiceActivityActive {
		return internal.ErrUnauthorized
	}
	_, err = repo.UpdateServiceRequest(ctx, repository.UpdateServiceRequestParams{
		StatusDetail: repository.ServiceRequestStatusCancelled,
		Activity:     repository.ServiceActivityInactive,
		ID:           requestID,
	})

	if err != nil {
		log.Printf("CancelServiceRequest: failed to update service request: %s\n", err)
		return internal.ErrInternalServerError
	}
	_, err = repo.AddTokens(ctx, repository.AddTokensParams{
		TokenBalance: repoRequest.SrTokenReward,
		ID:           repoRequest.RequesterID,
	})

	if err != nil {
		log.Printf("CancelServiceRequest: failed to add user tokens: %s\n", err)
		return internal.ErrInternalServerError
	}

	paymentID, err := repo.UpdatePaymentHolding(ctx, repository.UpdatePaymentHoldingParams{
		Status:           "refunded",
		ServiceRequestID: repoRequest.SrID,
	})
	if err != nil {
		log.Printf("CancelServiceRequest: failed to update payment holding: %s\n", err)
		return internal.ErrInternalServerError
	}

	err = repo.InsertTransaction(ctx, repository.InsertTransactionParams{
		UserID:    repoRequest.RequesterID,
		Type:      "addition",
		PaymentID: paymentID,
	})

	if err != nil {
		log.Printf("CancelServiceRequest: failed to insert transaction: %s\n", err)
		return internal.ErrInternalServerError
	}

	eventID, err := repo.InsertEvent(ctx, repository.InsertEventParams{
		TargetID:    requestID,
		Type:        domain.REQUEST_EVENT,
		Description: domain.CANCELLED_REQUEST,
	})
	if err != nil {
		log.Printf("CancelServiceRequest: failed to insert event: %s\n", err)
		return internal.ErrInternalServerError
	}
	_, err = repo.InsertNotification(ctx, repository.InsertNotificationParams{
		Message:         fmt.Sprintf("%s cancelled request for your service \"%s\".", repoRequest.RequesterFullName, repoRequest.SlTitle),
		RecipientUserID: repoRequest.ProviderID,
		ActionUserID:    pgtype.Text{String: repoRequest.RequesterID, Valid: true},
		EventID:         eventID,
	})
	if err != nil {
		log.Printf("CancelServiceRequest: failed to insert notification: %s\n", err)
		return internal.ErrInternalServerError
	}
	if err = tx.Commit(ctx); err != nil {
		log.Printf("CancelServiceRequest: failed to commit transaction: %s\n", err)
		return internal.ErrInternalServerError
	}
	err = internal.PusherClient.Trigger(fmt.Sprintf("user-%s", repoRequest.RequesterID), "new-notification", nil)
	if err != nil {
		log.Printf("CancelServiceRequest: failed to push notification: %s\n", err)
	}
	return nil
}

func (prs *PostgresRequestService) GetAllUserRequestReports(ctx context.Context, userID string) ([]RequestReport, error) {
	repo := repository.New(prs.DB)
	dbReports, err := repo.GetAllUserTickets(ctx, userID)
	if err != nil {
		log.Printf("GetAllUserTickets: failed to get all user tickets: %s\n", err)
		return nil, internal.ErrInternalServerError
	}
	tickets := make([]RequestReport, len(dbReports))
	for i, dbr := range dbReports {
		tickets[i] = RequestReport{
			ID:        dbr.ID,
			UserID:    dbr.ReporterID,
			RequestID: dbr.RequestID,
			TicketID:  dbr.TicketID,
			Status:    dbr.Status,
		}
	}

	return tickets, nil
}
