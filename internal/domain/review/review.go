package review

type Review struct {
	ID         int32  `json:"id"`
	RequestID  int32  `json:"request_id"`
	ReviewerID string `json:"reviewer_id"`
	RevieweeID string `json:"reviewee_id"`
	Comment    string `json:"comment"`
	Rating     int32  `json:"rating"`
}
