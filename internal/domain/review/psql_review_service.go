package review

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/domain"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type PostgresReviewService struct {
	DB *pgxpool.Pool
}

func (prs *PostgresReviewService) InsertReview(ctx context.Context, r Review) (int32, error) {
	tx, err := prs.DB.Begin(ctx)
	if err != nil {
		log.Println("InsertRequestReview: failed to being transaction: ", err)
		return -1, internal.ErrInternalServerError
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Printf("failed to rollback tx: %s\n", err)
		}
	}()
	repo := repository.New(prs.DB).WithTx(tx)
	insertedData, err := repo.InsertServiceRequestReview(ctx, repository.InsertServiceRequestReviewParams{
		RequestID:  r.RequestID,
		ReviewerID: r.ReviewerID,
		Comment:    pgtype.Text{String: r.Comment, Valid: !(r.Comment == "")},
		Rating:     r.Rating,
	})
	if err != nil {
		log.Println("InsertRequestReview: failed to insert review: ", err)
		return -1, internal.ErrInternalServerError
	}

	log.Println(r)

	_, err = repo.UpdateUserRating(ctx, repository.UpdateUserRatingParams{
		TotalRatings: r.Rating,
		RequestID:    r.RequestID,
	})

	if err != nil {
		log.Println("InsertRequestReview: failed to update user rating: ", err)
		return -1, internal.ErrInternalServerError
	}

	eventID, err := repo.InsertEvent(ctx, repository.InsertEventParams{
		TargetID: r.RequestID,
		Type:     domain.REVIEW_EVENT,
	})

	if err != nil {
		log.Println("InsertRequestReview: failed to insert event: ", err)
		return -1, internal.ErrInternalServerError
	}

	fullName, err := repo.GetUserFullNameByID(ctx, r.ReviewerID)
	if err != nil {
		log.Printf("InsertReview: failed to get user full_nam:%s\n", err)
		return -1, internal.ErrInternalServerError
	}
	_, err = repo.InsertNotification(ctx, repository.InsertNotificationParams{
		Message:         fmt.Sprintf("%s left a review on you.", fullName),
		RecipientUserID: insertedData.RevieweeID,
		ActionUserID:    pgtype.Text{String: r.ReviewerID, Valid: true},
		EventID:         eventID,
	})

	if err != nil {
		log.Println("InsertRequestReview: failed to insert notification: ", err)
		return -1, internal.ErrInternalServerError
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("InsertRequestReview: failed to commit transaction: %s\n", err)
		return -1, internal.ErrInternalServerError
	}
	err = internal.PusherClient.Trigger(fmt.Sprintf("user-%s", insertedData.RevieweeID), "new-notifications", nil)
	if err != nil {
		log.Printf("failed to trigger pusher notification: %s\n", err)
	}
	return insertedData.ID, nil
}

func (prs *PostgresReviewService) GetReviewByID(ctx context.Context, reviewID int32) (Review, error) {
	repo := repository.New(prs.DB)
	var r Review
	dbReview, err := repo.GetReviewByID(ctx, reviewID)
	if err != nil {
		log.Printf("GetReviewByID: failed to get review by id: %d\n", err)
		return r, internal.ErrInternalServerError
	}
	r.ID = dbReview.ID
	r.Rating = dbReview.Rating
	r.RequestID = dbReview.RequestID
	r.RevieweeID = dbReview.RevieweeID
	r.ReviewerID = dbReview.ReviewerID
	r.Comment = dbReview.Comment.String
	r.CreatedAt = dbReview.DateTime
	return r, nil
}
