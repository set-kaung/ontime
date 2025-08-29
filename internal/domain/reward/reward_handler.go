package reward

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/helpers"
)

type RewardHandler struct {
	RewardService RewardService
}

func (rh *RewardHandler) HandleGetAllRewards(w http.ResponseWriter, r *http.Request) {
	rewards, err := rh.RewardService.GetAllRewards(r.Context())
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, rewards, nil)
}

func (rh *RewardHandler) HandleGetAllUserRedeemdRewards(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	redeemedRewards, err := rh.RewardService.GetAllUserRedeemedRewards(r.Context(), userID)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, redeemedRewards, nil)
}

func (rh *RewardHandler) HandleRewardByID(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	rewardID, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid reward id", nil)
		return
	}
	reward, err := rh.RewardService.GetRewardByID(r.Context(), int32(rewardID))
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, reward, nil)
}

func (rh *RewardHandler) HandleRedeemReward(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	pathID := r.PathValue("id")
	rewardID, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid reward id", nil)
		return
	}
	couponCode, err := rh.RewardService.InsertRedeemedReward(r.Context(), int32(rewardID), userID)
	if err != nil {
		if !errors.Is(err, internal.ErrMismatchAmount) {
			helpers.WriteServerError(w, nil)
			return
		}
		if errors.Is(err, internal.ErrInsufficientBalance) {
			helpers.WriteError(w, http.StatusBadRequest, internal.ErrInsufficientBalance.Error(), nil)
			return
		}
		helpers.WriteError(w, http.StatusBadRequest, "invalid rewards", nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, map[string]string{"coupon_code": couponCode}, nil)
}
