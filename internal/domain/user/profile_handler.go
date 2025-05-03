package user

import (
	"log"
	"net/http"

	"github.com/set-kaung/senior_project_1/internal/helpers"
)

func (uh *UserHandler) ViewOwnProfile(w http.ResponseWriter, r *http.Request) {
	userID := uh.SessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 {
		helpers.WriteError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	user, err := uh.UserService.GetUserProfile(r.Context(), int32(userID))
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, "server is having problems", nil)
		return
	}
	err = helpers.WriteData(w, http.StatusOK, user, nil)
	if err != nil {
		log.Println("error writing data:", err)
	}
}
