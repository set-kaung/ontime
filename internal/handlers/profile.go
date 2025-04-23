package handlers

import (
	"net/http"

	"github.com/set-kaung/senior_project_1/internal/helpers"
)

func (uh *UserHandler) ViewOwnProfile(w http.ResponseWriter, r *http.Request) {
	userID := uh.SessionManager.GetInt(r.Context(), "authenticatedUserID")
	if userID == 0 {
		helpers.WriteError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

}
