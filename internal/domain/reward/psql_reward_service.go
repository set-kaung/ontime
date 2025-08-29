package reward

import (
	"context"
	"log"
	"strings"

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
			RewardTitle:        rr.Title.String,
			RedeemedAt:         rr.RedeemedAt,
			RedeemedUserID:     rr.UserID,
			CostAtRedeemedTime: rr.RedeemedCost,
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

func (prs *PostgresRewardService) InsertRedeemedReward(ctx context.Context, rewardID int32, userID string) (string, error) {
	tx, err := prs.DB.Begin(ctx)
	if err != nil {
		log.Printf("InsertRedeemedReward: failed to start transaction: %v\n", err)
		return "", internal.ErrInternalServerError
	}
	defer tx.Rollback(ctx)
	repo := repository.New(prs.DB).WithTx(tx)
	couponCode, err := repo.InsertRedeemedReward(ctx, repository.InsertRedeemedRewardParams{
		RewardID: rewardID,
		UserID:   userID,
	})
	if err != nil {
		log.Printf("InsertRedeemedReward: failed to insert redeemed reward: %v\n", err)
		return "", internal.ErrInternalServerError
	}
	if strings.TrimSpace(couponCode) == "" {
		log.Printf("InsertRedeemedReward: failed to insert redeemed reward: insufficient balance\n")
		return "", internal.ErrInsufficientBalance
	}
	rowsAffected, err := repo.DeductRewardAmount(ctx, rewardID)
	if err != nil {
		log.Printf("InsertRedeemedReward: failed to deduct reward amount: %v\n", err)
		return "", internal.ErrInternalServerError
	}
	if rowsAffected != 1 {
		log.Printf("InsertRedeemedReward: failed to deduct reward amount: invalid.\n")
		return "", internal.ErrMismatchAmount
	}
	err = repo.DeductRewardTokens(ctx, repository.DeductRewardTokensParams{
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
	return couponCode, nil
}
