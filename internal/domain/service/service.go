package domain

import "time"

type Service struct {
	UserID      int
	Title       string
	Description string
	Cost        int
	Date        time.Time
}
