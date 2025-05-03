package user

import "net/mail"

type Email struct {
	Address string `json:"email"`
}

func NewEmail(email string) (Email, error) {
	_, err := mail.ParseAddress(email)
	return Email{email}, err
}
