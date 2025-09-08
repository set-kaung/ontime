package review

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/helpers"
)

type ReviewHandler struct {
	ReviewService ReviewService
}

func (rh *ReviewHandler) HandleSubmitReview(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	pathID := r.PathValue("id")
	requestID, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		log.Println("HandleSubmitReview: err: ", err)
		helpers.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	review := Review{}
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&review)
	review.ReviewerID = userID
	review.RequestID = int32(requestID)

	fmt.Println("From user: ", review)

	reviewID, err := rh.ReviewService.InsertReview(r.Context(), review)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusCreated, map[string]int32{"review_id": reviewID}, nil)
}

func (rh *ReviewHandler) HandleGetReviewByID(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	reviewID, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		log.Println("HandleGetReviewByID: err: ", err)
		helpers.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}
	review, err := rh.ReviewService.GetReviewByID(r.Context(), int32(reviewID))
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, review, nil)
}
