package user

import "net/mail"

type User struct {
	ID           int    `json:"ID"`
	Email        Email  `json:"email"`
	PasswordHash string `json:"passwordHash"`
	Profile      `json:"profile"`
}

type Email struct {
	Address string `json:"email"`
}

func NewEmail(email string) (Email, error) {
	_, err := mail.ParseAddress(email)
	return Email{email}, err
}
