package user

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/helpers"
)

type UserHandler struct {
	UserService UserService
}

func (h *UserHandler) HandleInsertUser(w http.ResponseWriter, r *http.Request) {
	claims, ok := clerk.SessionClaimsFromContext(r.Context())
	if !ok {
		log.Println("failed to session claims")
		helpers.WriteError(w, http.StatusInternalServerError, internal.ErrInternalServerError.Error(), nil)
	}
	clerkUser, err := user.Get(r.Context(), claims.Subject)
	if err != nil {
		log.Println("authenticate -> GetClerkUserID: ", err)

	}
	id, err := GetClerkUserID(r.Context())
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	dbUser := User{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&dbUser)
	if err != nil {
		log.Println("user_handler -> HandleInsertUser: ", err)
		helpers.WriteError(w, http.StatusBadRequest, "invalid json request", nil)
		return
	}
	dbUser.ID = id
	dbUser.FirstName = *clerkUser.FirstName
	dbUser.LastName = *clerkUser.LastName
	dbUser.Status = "active"
	err = h.UserService.InsertUser(r.Context(), dbUser)
	if err != nil {
		log.Println("user_handler -> HandleInsertUser: ", err)
		helpers.WriteError(w, http.StatusInternalServerError, internal.ErrInternalServerError.Error(), nil)
		return
	}
	// need to update clerk public metadata for profileCompletion
	updateDate := map[string]interface{}{
		"profileComplete": true,
	}

	payload, err := json.Marshal(updateDate)
	if err != nil {
		panic("user_handler -> HandleInsertUser: invalid json")
	}
	_, err = user.Update(r.Context(), claims.Subject, &user.UpdateParams{
		PublicMetadata: clerk.JSONRawMessage(payload),
	})
	if err != nil {
		log.Println("failed to update Clerk user metadata:", err)
	}
}

func (h *UserHandler) HandleViewOwnProfile(w http.ResponseWriter, r *http.Request) {

	id, err := GetClerkUserID(r.Context())
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	dbUser, err := h.UserService.GetUserByID(r.Context(), id)
	if err != nil {
		log.Println("user_handler -> HandleViewOwnProfile: ", err)
		helpers.WriteError(w, http.StatusInternalServerError, "no user data", nil)
		return
	}
	err = helpers.WriteData(w, http.StatusOK, dbUser, nil)
	if err != nil {
		log.Println("user_handler -> HandleViewOwnProfile: ", err)
	}
}
