package reward

import "context"

type RewardService interface {
	GetAllRewards(context.Context) ([]Reward, error)
	GetAllUserRedeemedRewards(context.Context, string) ([]RedeemedReward, error)
	GetRewardByID(context.Context, int32) (Reward, error)
	InsertRedeemedReward(ctx context.Context, rewardID int32, userID string) (string, error)
}
