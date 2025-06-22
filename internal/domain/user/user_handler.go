package user

import (
	"encoding/json"
	"errors"
	"fmt"
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
		log.Println(err)
		helpers.WriteError(w, http.StatusUnauthorized, err.Error(), nil)
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

func (h *UserHandler) HandleUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	id, err := GetClerkUserID(r.Context())
	if err != nil {
		if errors.Is(err, internal.ErrUnauthorized) {
			helpers.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		helpers.WriteServerError(w, nil)
		return
	}
	decoder := json.NewDecoder(r.Body)
	u := User{}
	err = decoder.Decode(&u)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	u.ID = id
	err = h.UserService.UpdateUser(r.Context(), u)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteSuccess(w, http.StatusOK, "user data updated", nil)
}

func (uh *UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := GetClerkUserID(r.Context())
	if err != nil {
		if errors.Is(err, internal.ErrUnauthorized) {
			helpers.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
			return
		}
		helpers.WriteServerError(w, nil)
		return
	}
	deletedResource, err := user.Delete(r.Context(), id)
	if err != nil {
		log.Printf("UserHandler -> HandleDeleteUser: error deleting clerk user: %v", err)
		helpers.WriteServerError(w, nil)
		return
	}
	fmt.Printf("User with ID %s deleted successfully. Deleted resource ID: %s\n", id, deletedResource.ID)
	err = uh.UserService.DeleteUser(r.Context(), id)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}

}
