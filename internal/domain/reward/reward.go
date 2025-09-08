package reward

import "time"

type Reward struct {
	ID              int32     `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Cost            int32     `json:"cost"`
	AvailableAmount int64     `json:"available_amount"`
	ImageURL        string    `json:"image_url"`
	CreatedDate     time.Time `json:"created_date"`
}

type RedeemedReward struct {
	ID                 int32     `json:"id"`
	RewardID           int32     `json:"reward_id"`
	RewardTitle        string    `json:"reward_title"`
	RewardDescription  string    `json:"description"`
	RedeemedAt         time.Time `json:"redeemed_at"`
	RedeemedUserID     string    `json:"reedemed_user_id"`
	CostAtRedeemedTime int32     `json:"cost_at_redeemed_time"`
	ImageURL           string    `json:"image_url"`
	CouponCode         string    `json:"coupon_code"`
}
