package user

import (
	"log"

	"github.com/set-kaung/senior_project_1/internal/util"
)

type AuthenticationService struct {
	Repo UserRepository
}

func (as *AuthenticationService) Authenticate(email string, password string) (int, bool) {
	user, err := as.Repo.GetUserByEmail(email)
	if err != nil {
		log.Println(err)
		return -1, false
	}
	return user.ID, util.CheckPasswordHash(password, user.PasswordHash)
}
