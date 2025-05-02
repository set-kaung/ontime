package user

import "time"

type Profile struct {
	Username string `json:"username"`
	Tokens   int    `json:"tokens"`
	Email    Email  `json:"email"`
	// Rating   float32 `json:"Rating"`
	Joined time.Time `json:"joined"`
}
