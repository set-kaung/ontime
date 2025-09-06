package review

type Review struct {
	ID               int32  `json:"id"`
	RequestID        int32  `json:"request_id"`
	ReviewerID       string `json:"reviewer_id"`
	ReviewerFullName string `json:"reviwer_full_name"`
	RevieweeFullName string `json:"reviwee_full_name"`
	RevieweeID       string `json:"reviewee_id"`
	Comment          string `json:"comment"`
	Rating           int32  `json:"rating"`
}
