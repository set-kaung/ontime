package reward

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type PostgresRewardService struct {
	DB *pgxpool.Pool
}

func (prs *PostgresRewardService) GetAllRewards(ctx context.Context) ([]Reward, error) {
	var rewards []Reward
	repo := repository.New(prs.DB)
	repoRewards, err := repo.GetAllRewards(ctx)
	if err != nil {
		log.Printf("GetAllRewards: failed to get all rewards: %v\n", err)
		return nil, internal.ErrInternalServerError
	}
	rewards = make([]Reward, len(repoRewards))
	for i, r := range repoRewards {
		rewards[i] = Reward{
			ID:              r.ID,
			Title:           r.Title,
			Description:     r.Description,
			Cost:            r.Cost,
			AvailableAmount: r.AvailableAmount,
			ImageURL:        r.ImageUrl.String,
			CreatedDate:     r.CreatedDate,
		}
	}
	return rewards, nil
}

func (prs *PostgresRewardService) GetAllUserRedeemedRewards(ctx context.Context, userID string) ([]RedeemedReward, error) {
	var redeemedRewards []RedeemedReward
	repo := repository.New(prs.DB)
	repoRedeemed, err := repo.GetAllUserRedeemdRewards(ctx, userID)
	if err != nil {
		log.Printf("GetAllUserRedeemedRewards: failed to get all reedeemed rewards: %v\n", err)
		return nil, internal.ErrInternalServerError
	}
	redeemedRewards = make([]RedeemedReward, len(repoRedeemed))
	for i, rr := range repoRedeemed {
		redeemedRewards[i] = RedeemedReward{
			RewardID:           rr.RewardID,
			RewardTitle:        rr.Title,
			RewardDescription:  rr.Description,
			RedeemedAt:         rr.RedeemedAt,
			RedeemedUserID:     rr.UserID,
			CostAtRedeemedTime: rr.RedeemedCost,
			ImageURL:           rr.ImageUrl.String,
			CouponCode:         rr.CouponCode,
		}
	}
	return redeemedRewards, err
}

func (prs *PostgresRewardService) GetRewardByID(ctx context.Context, rewardID int32) (Reward, error) {
	var reward Reward
	repo := repository.New(prs.DB)
	r, err := repo.GetRewardByID(ctx, rewardID)
	if err != nil {
		log.Printf("GetRewardByID: failed to get all reedeemed rewards: %v\n", err)
		return reward, err
	}

	reward = Reward{
		ID:              r.ID,
		Title:           r.Title,
		Description:     r.Description,
		Cost:            r.Cost,
		AvailableAmount: r.AvailableAmount,
		ImageURL:        r.ImageUrl.String,
		CreatedDate:     r.CreatedDate,
	}
	return reward, nil
}

// 1. Get all coupon codes
// 2. Choose one at random and update the status
// 3. If successful, insert RedeemedReward
// 4. Deduct users tokens
// 5. Return coupon code
func (prs *PostgresRewardService) InsertRedeemedReward(ctx context.Context, rewardID int32, userID string) (string, error) {
	tx, err := prs.DB.Begin(ctx)
	if err != nil {
		log.Printf("InsertRedeemedReward: failed to start transaction: %v\n", err)
		return "", internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)

	repo := repository.New(prs.DB).WithTx(tx)

	couponCodes, err := repo.GetAllCouponCodes(ctx, rewardID)
	if err != nil {
		log.Printf("InsertRedeemedReward: failed to get reward's coupons: %v\n", err)
		return "", internal.ErrInternalServerError
	}
	if len(couponCodes) < 1 {
		log.Printf("InsertRedeemedReward: fatal: invalid couponCodes len %v\n")
		return "", internal.ErrInternalServerError
	}
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	coupon := couponCodes[rand.Intn(len(couponCodes))]

	rows, err := repo.UpdateCouponCodeStatus(ctx, coupon.ID)
	if err != nil {
		log.Printf("InsertRedeemedReward: failed to update coupon code status: %v\n", err)
		return "", internal.ErrInternalServerError
	}

	if rows != 1 {
		log.Printf("InsertRedeemedReward: failed to update only one coupon code: %d rows affected\n", rows)
		return "", internal.ErrInternalServerError
	}
	rows, err = repo.InsertRedeemedReward(ctx, repository.InsertRedeemedRewardParams{
		RewardID:     rewardID,
		UserID:       userID,
		CouponCodeID: coupon.ID,
	})
	if err != nil {
		log.Printf("InsertRedeemedReward: failed to insert redeemed reward: %v\n", err)
		return "", internal.ErrInternalServerError
	}
	if rows != 1 {
		log.Printf("InsertRedeemedReward: failed to insert redeemed reward: insufficient balance\n")
		return "", internal.ErrInsufficientBalance
	}

	err = repo.DeductRewardTokensFromUser(ctx, repository.DeductRewardTokensFromUserParams{
		UserID:   userID,
		RewardID: rewardID,
	})
	if err != nil {
		log.Printf("InsertRedeemedReward: failed to deduct reward cost from user: %v\n", err)
		return "", internal.ErrInternalServerError
	}
	if err = tx.Commit(ctx); err != nil {
		log.Printf("InsertRedeemedReward: failed to commit: %v\n", err)
		return "", internal.ErrInternalServerError
	}
	return coupon.CouponCode, nil
}

func (p *PostgresRewardService) GetRedeemedRewardByID(ctx context.Context, redeemedRewardID int32) (RedeemedReward, error) {
	repo := repository.New(p.DB)
	dbRR, err := repo.GetRedeemedRewardByID(ctx, redeemedRewardID)
	var rr RedeemedReward
	if err != nil {
		log.Printf("GetRedeemedRewardByID: failed to get redeemed reward: %s\n", err)
		return rr, internal.ErrInternalServerError
	}
	rr.ID = dbRR.RedeemedID
	rr.CostAtRedeemedTime = dbRR.Cost
	rr.RedeemedAt = dbRR.RedeemedAt
	rr.RewardID = dbRR.RewardID
	rr.RedeemedUserID = dbRR.UserID
	rr.RewardTitle = dbRR.Title
	rr.RewardDescription = dbRR.Description
	rr.ImageURL = dbRR.ImageUrl.String
	rr.CouponCode = dbRR.CouponCode
	return rr, nil
}
