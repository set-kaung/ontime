package user

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/domain"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type PostgresUserService struct {
	DB *pgxpool.Pool
}

func (pus *PostgresUserService) GetUserByID(ctx context.Context, id string) (User, error) {
	repo := repository.New(pus.DB)
	repoUser, err := repo.GetUserByID(ctx, id)

	user := User{}
	user.ID = repoUser.ID
	user.FullName = repoUser.FullName
	user.TokenBalance = repoUser.TokenBalance
	user.Status = string(repoUser.Status)
	user.AddressLine1 = repoUser.AddressLine1
	user.AddressLine2 = repoUser.AddressLine2
	user.City = repoUser.City
	user.StateProvince = repoUser.StateProvince
	user.ZipPostalCode = repoUser.ZipPostalCode
	user.Country = repoUser.Country
	user.JoinedAt = repoUser.JoinedAt
	user.IsEmailSignedUp = repoUser.IsEmailSignedup
	user.ServicesProvided = uint32(repoUser.ServicesProvided)
	user.ServicesReceived = uint32(repoUser.ServicesReceived)
	rating := float32(repoUser.TotalRatings.Int32) / max(1.0, float32(repoUser.RatingCount.Int32))
	user.Rating = float32(math.Round(float64(rating)*100) / 100)
	user.AboutMe = repoUser.AboutMe.String
	return user, err
}

func (pus *PostgresUserService) InsertUser(ctx context.Context, user User) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Printf("failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	insertUserParams := repository.InsertUserParams{
		ID:            user.ID,
		FullName:      user.FullName,
		Phone:         user.Phone,
		TokenBalance:  user.TokenBalance,
		AddressLine1:  user.AddressLine1,
		AddressLine2:  user.AddressLine2,
		City:          user.City,
		StateProvince: user.StateProvince,
		ZipPostalCode: user.ZipPostalCode,
		Country:       user.Country,
	}
	insertUserParams.Status = repository.AccountStatusActive

	repo := repository.New(pus.DB).WithTx(tx)
	_, err = repo.InsertUser(ctx, insertUserParams)
	if err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok {
			if pgerr.Code == "23505" {
				log.Printf("InsertUser: failed to insert user: %v\n", err)
				return internal.ErrDuplicateID
			}
		}
		log.Printf("InsertUser: failed to insert user: %v\n", err)
		return internal.ErrInternalServerError
	}
	err = repo.InsertNewUserRating(ctx, user.ID)
	if err != nil {
		log.Printf("UserService -> InsertUser: error adding user ratings: %s\n", err)
		return internal.ErrInternalServerError
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("UserService -> InsertUser: error commiting transaction: %s\n", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (pus *PostgresUserService) UpdateUser(ctx context.Context, user User) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Printf("failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	updateUserParams := repository.UpdateUserParams{
		FullName:      user.FullName,
		Phone:         user.Phone,
		AddressLine1:  user.AddressLine1,
		AddressLine2:  user.AddressLine2,
		City:          user.City,
		StateProvince: user.StateProvince,
		Country:       user.Country,
		ZipPostalCode: user.ZipPostalCode,
		ID:            user.ID,
	}
	repo := repository.New(pus.DB).WithTx(tx)
	_, err = repo.UpdateUser(ctx, updateUserParams)
	if err != nil {
		log.Printf("UserService -> UpdateUser: error: %s\n", err)
		return internal.ErrInternalServerError
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("UserService -> InsertUser: error commiting transaction: %s\n", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (pus *PostgresUserService) DeleteUser(ctx context.Context, id string) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Printf("failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Printf("DeleteUser: failed to rollback tx: %s\n", err)
		}
	}()

	repo := repository.New(pus.DB).WithTx(tx)
	requestIDs, err := repo.GetProvidingeRequests(ctx, id)
	if err != nil {
		log.Printf("DeleteUser: failed to get requestIDs: %s\n", err)
		return internal.ErrInternalServerError
	}
	for _, request := range requestIDs {

		payment, err := repo.GetRequestPayment(ctx, request.ID)
		if err != nil {
			log.Printf("DeleteUser: failed to get request payment: %s\n", err)
			return internal.ErrInternalServerError
		}
		_, err = repo.AddTokens(ctx, repository.AddTokensParams{
			ID:           payment.PayerID,
			TokenBalance: payment.AmountTokens,
		})
		if err != nil {
			log.Printf("DeleteUser: failed to add tokens: %s\n", err)
			return internal.ErrInternalServerError
		}
		_, err = repo.UpdatePaymentHolding(ctx, repository.UpdatePaymentHoldingParams{
			Status:           repository.PaymentStatusRefunded,
			ServiceRequestID: request.ID,
		})
		if err != nil {
			log.Printf("DeleteUser: failed to update payment holdings: %s\n", err)
			return internal.ErrInternalServerError
		}
		err = repo.InsertTransaction(ctx, repository.InsertTransactionParams{
			UserID:    payment.PayerID,
			Type:      "addition",
			PaymentID: payment.ID,
		})
		if err != nil {
			log.Printf("DeleteUser: failed to insert transaction: %s\n", err)
			return internal.ErrInternalServerError
		}
		eventID, err := repo.InsertEvent(ctx, repository.InsertEventParams{
			TargetID:    request.ID,
			Type:        domain.SYSTEM_EVENT,
			Description: domain.USER_DO_NOT_EXIST,
		})
		if err != nil {
			log.Printf("DeleteUser: failed to  insert event: %s\n", err)
			return internal.ErrInternalServerError
		}
		_, err = repo.InsertNotification(ctx, repository.InsertNotificationParams{
			Message:         fmt.Sprintf("Your token for %s has been refunded.", request.Title),
			RecipientUserID: payment.PayerID,
			ActionUserID:    pgtype.Text{Valid: false},
			EventID:         eventID,
		})
		if err != nil {
			log.Printf("DeleteUser: failed to insert notification: %s\n", err)
			return internal.ErrInternalServerError
		}
		err = internal.PusherClient.Trigger(fmt.Sprintf("user-%s", payment.PayerID), "new-notification", nil)
		if err != nil {
			log.Printf("DeleteUser: failed to trigger pusher: %s\n", err)
		}
	}

	_, err = repo.DeleteUser(ctx, id)
	if err != nil {
		log.Printf("UserService -> DeleteUser: error: %s\n", err)
		return internal.ErrInternalServerError
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Printf("UserService -> DeleteUser: error commiting transaction: %s\n", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (pus *PostgresUserService) InsertAdsHistory(ctx context.Context, userID string) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Printf("failed to begin tx: %s\n", err)
		return internal.ErrInternalServerError
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Printf("failed to rollback tx: %s\n", err)
		}
	}()

	repo := repository.New(pus.DB).WithTx(tx)
	_, err = repo.InsertAdsHistory(ctx, userID)
	if err != nil {
		log.Printf("failed to insert ads history: %s\n", err)
		return internal.ErrInternalServerError
	}
	_, err = repo.AddTokens(ctx, repository.AddTokensParams{
		TokenBalance: 1,
		ID:           userID,
	})
	if err != nil {
		log.Printf("failed to add token balance for ad watching: %s\n", err)
		return internal.ErrInternalServerError
	}
	if err = tx.Commit(ctx); err != nil {
		log.Printf("failed to commit ads history: %s\n", err)
		return internal.ErrInternalServerError
	}

	return nil
}

func (pus *PostgresUserService) GetAdsHistory(ctx context.Context, userID string) (int64, error) {
	repo := repository.New(pus.DB)
	count, err := repo.GetAdsWatched(ctx, userID)
	if err != nil {
		log.Printf("failed to get ads history: %s\n", err)
		return -1, internal.ErrInternalServerError
	}
	return count, nil
}

func (pus *PostgresUserService) GetNotifications(ctx context.Context, userID string) ([]Notification, error) {
	repo := repository.New(pus.DB)
	dbNotis, err := repo.GetNotifications(ctx, userID)
	if err != nil {
		log.Printf("failed to get notifications: %s\n", err)
		return nil, internal.ErrInternalServerError
	}
	notifications := make([]Notification, 0, len(dbNotis))
	for _, dbN := range dbNotis {
		notifications = append(notifications, Notification{
			ID:              dbN.ID,
			ActionUserID:    dbN.ActionUserID.String,
			RecipientUserID: dbN.RecipientUserID,
			EventID:         dbN.EventID,
			Message:         dbN.Message,
			IsRead:          dbN.IsRead,
			CreatedAt:       dbN.CreatedAt,
			EventType:       dbN.Type,
			EventTargetID:   dbN.TargetID,
		})
	}

	return notifications, nil
}

func (pus *PostgresUserService) UpdateNotificationStatus(ctx context.Context, userID string, notiID int32) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Println("UpdateNotificationStatus: failed to start transaction: ", err)
		return internal.ErrInternalServerError
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Printf("failed to rollback tx: %s\n", err)
		}
	}()
	repo := repository.New(pus.DB).WithTx(tx)
	_, err = repo.SetUserNotificationsRead(ctx, repository.SetUserNotificationsReadParams{
		RecipientUserID: userID,
		ID:              notiID,
	})
	if err != nil {
		log.Println("UpdateNotificationStatus: failed to update notifications: ", err)
		return internal.ErrInternalServerError
	}
	if err = tx.Commit(ctx); err != nil {
		log.Println("UpdateNotificationStatus: failed to commit transaction: ", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (pus *PostgresUserService) MarkAllNotificationsRead(ctx context.Context, recipientUserID string, targetTime time.Time) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Println("MarkAllAllNotificationsRead: failed to start transaction: ", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(pus.DB).WithTx(tx)
	err = repo.SetAllNotificationsRead(ctx, repository.SetAllNotificationsReadParams{
		RecipientUserID: recipientUserID,
		CreatedAt:       targetTime,
	})

	if err != nil {
		log.Println("MarkAllAllNotificationsRead: failed to update all notifications: ", err)
		return internal.ErrInternalServerError
	}

	if err = tx.Commit(ctx); err != nil {
		log.Println("MarkAllNotificationsRead: failed to commit transaction: ", err)
	}
	return nil
}

func (pus *PostgresUserService) UpdateFullName(ctx context.Context, newName string, userID string) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Println("UpdateUserFullName: failed to start transactions: ", err)
		return internal.ErrInternalServerError
	}
	repo := repository.New(pus.DB).WithTx(tx)
	defer tx.Rollback(ctx)

	_, err = repo.UpdateUserFullNmae(ctx, repository.UpdateUserFullNmaeParams{
		FullName: newName,
		ID:       userID,
	})

	if err != nil {
		log.Println("UpdateUserFullName: failed to update user full name: ", err)
		return internal.ErrInternalServerError
	}
	if err = tx.Commit(ctx); err != nil {
		log.Println("UpdateUserFullName: failed to commit: ", err)
		return internal.ErrInternalServerError
	}
	return nil
}

func (pus *PostgresUserService) GetAllHistory(ctx context.Context, userID string) ([]InteractionHistory, error) {
	var interactionHistories []InteractionHistory
	repo := repository.New(pus.DB)
	repoRequests, err := repo.GetAllUserRequests(ctx, userID)
	if err != nil {
		log.Printf("GetAllHistory: failed to get request histories: %v\n", err)
		return nil, internal.ErrInternalServerError
	}

	for _, r := range repoRequests {
		interactionHistories = append(interactionHistories, InteractionHistory{
			InteractionType: "request",
			Description:     fmt.Sprintf("Service Request: %s", r.Title),
			IsIncoming:      r.ProviderID == userID,
			TargetID:        r.ID,
			Amount:          r.TokenReward,
			Status:          string(r.StatusDetail),
			Timestamp:       r.CreatedAt,
		})
	}

	adsHistories, err := repo.GetAdsHistory(ctx, userID)
	if err != nil {
		log.Printf("GetAllHistory: failed to get ads histories: %v\n", err)
		return nil, internal.ErrInternalServerError
	}
	for _, a := range adsHistories {
		interactionHistories = append(interactionHistories, InteractionHistory{
			InteractionType: "advertisement",
			Description:     "Advertisement Watched",
			IsIncoming:      true,
			TargetID:        a.ID,
			Amount:          1,
			Status:          "completed",
			Timestamp:       a.DateTime,
		})
	}

	rewardHistories, err := repo.GetAllUserRedeemdRewards(ctx, userID)
	if err != nil {
		log.Printf("GetAllHistory: failed to get redeemed reward histories: %v\n", err)
		return nil, internal.ErrInternalServerError
	}
	for _, rr := range rewardHistories {
		interactionHistories = append(interactionHistories, InteractionHistory{
			InteractionType: "reward",
			Description:     "Reward Redeemed",
			IsIncoming:      false,
			TargetID:        rr.ID,
			Amount:          rr.RedeemedCost,
			Status:          "completed",
			Timestamp:       rr.RedeemedAt,
		})
	}
	return interactionHistories, nil
}

func (pus *PostgresUserService) UpdateOneTimePaid(ctx context.Context, userID string) (int32, error) {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Printf("UpdateOneTimePaid: failed to create transaction: %v\n", err)
		return -1, internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(pus.DB).WithTx(tx)

	bonusStr := os.Getenv("ONETIME_PAYMENT_TOKENS")
	amount64, err := strconv.ParseInt(bonusStr, 10, 32)
	if err != nil {
		log.Printf("UpdateOneTimePaid: failed to parse env string: %v\n", err)
		return -1, internal.ErrInternalServerError
	}
	bonus := int32(amount64)

	newBalance, err := repo.MarkSignupPaidAndAward(ctx, repository.MarkSignupPaidAndAwardParams{
		ID:           userID,
		TokenBalance: bonus,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, nil
		}
		log.Printf("UpdateOneTimePaid: failed to marksignup %v\n", err)
		return -1, internal.ErrInternalServerError
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("UpdateOneTimePaid: failed to commit %v\n", err)
		return -1, internal.ErrInternalServerError
	}
	return newBalance, nil
}

func (pus *PostgresUserService) GetUserDetailAndServices(ctx context.Context, userID string) (UserSummary, error) {
	repo := repository.New(pus.DB)
	user, err := pus.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("GetUserDetailAndServices: failed to get user by id: %v\n", err)
		return UserSummary{}, internal.ErrInternalServerError
	}
	dbListings, err := repo.GetPartialListingsByUserID(ctx, userID)
	if err != nil {
		log.Printf("GetUserDetailAndServices: failed to get partial listings: %v\n", err)
		return UserSummary{}, internal.ErrInternalServerError
	}
	userSummary := UserSummary{User: user, Listings: make([]PartialListing, len(dbListings))}
	for i, dbListing := range dbListings {
		rating := float32(0)
		if dbListing.RatingCount != 0 {
			rating = float32(dbListing.TotalRating) / float32(dbListing.RatingCount)
		}
		userSummary.Listings[i] = PartialListing{
			ID:          dbListing.ID,
			Title:       dbListing.Title,
			Category:    dbListing.Category,
			AvgRating:   rating,
			RatingCount: int32(dbListing.RatingCount),
			TokenReward: dbListing.TokenReward,
			ImageURL:    dbListing.ImageUrl.String,
		}
	}
	return userSummary, nil
}

func (pus *PostgresUserService) UpdateUserAboutMe(ctx context.Context, userID string, aboutMe string) error {
	tx, err := pus.DB.Begin(ctx)
	if err != nil {
		log.Printf("UpdateUserAboutMe: failed to create transaction: %v\n", err)
		return internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(pus.DB).WithTx(tx)
	err = repo.UpdateAboutMe(ctx, repository.UpdateAboutMeParams{
		ID:      userID,
		AboutMe: pgtype.Text{String: aboutMe, Valid: true},
	})
	if err != nil {
		log.Printf("UpdateUserAboutMe: failed to update about me: %s\n", err)
		return internal.ErrInternalServerError
	}

	if err := tx.Commit(ctx); err != nil {
		log.Printf("UpdateUserAboutMe: failed to commit %v\n", err)
		return internal.ErrInternalServerError
	}
	return nil
}
