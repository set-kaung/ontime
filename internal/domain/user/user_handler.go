package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	dbUser := User{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&dbUser)
	if err != nil {
		log.Println("user_handler -> HandleInsertUser: ", err)
		helpers.WriteError(w, http.StatusBadRequest, "invalid json request", nil)
		return
	}
	dbUser.ID = userID
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

	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	dbUser, err := h.UserService.GetUserByID(r.Context(), userID)
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
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	decoder := json.NewDecoder(r.Body)
	u := User{}
	err := decoder.Decode(&u)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	u.ID = userID
	err = h.UserService.UpdateUser(r.Context(), u)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteSuccess(w, http.StatusOK, "user data updated", nil)
}

func (uh *UserHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	deletedResource, err := user.Delete(r.Context(), userID)
	if err != nil {
		log.Printf("UserHandler -> HandleDeleteUser: error deleting clerk user: %v", err)
		helpers.WriteServerError(w, nil)
		return
	}
	log.Printf("User with ID %s deleted successfully. Deleted resource ID: %s\n", userID, deletedResource.ID)
	err = uh.UserService.DeleteUser(r.Context(), userID)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteSuccess(w, http.StatusOK, "user deleted", nil)
}

func (uh *UserHandler) HandleAdWatched(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	err := uh.UserService.InsertAdsHistory(r.Context(), userID)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteSuccess(w, http.StatusOK, "ad watched", nil)
}

func (uh *UserHandler) HandleGetAdsWatched(w http.ResponseWriter, r *http.Request) {
	type AdsStatus struct {
		Count     int64 `json:"count"`
		IsAtLimit bool  `json:"is_at_limit"`
	}
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	count, err := uh.UserService.GetAdsHistory(r.Context(), userID)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, AdsStatus{
		Count:     count,
		IsAtLimit: fmt.Sprintf("%d", count) == os.Getenv("DAILY_ADS_LIMIT"),
	}, nil)
}

func (uh *UserHandler) GetUserNotifications(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	notifications, err := uh.UserService.GetNotifications(r.Context(), userID)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}

	helpers.WriteData(w, http.StatusOK, notifications, nil)
}

func (uh *UserHandler) HandleUpdateNotificationStatus(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)

	notiIDStr := r.URL.Query().Get("id")
	notiID, err := strconv.ParseInt(notiIDStr, 10, 32)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid notification id", nil)
		return
	}

	err = uh.UserService.UpdateNotificationStatus(r.Context(), userID, int32(notiID))
	if err != nil {
		helpers.WriteServerError(w, nil)
	}
}

func (uh *UserHandler) HandlUpdateUserFullName(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	data := map[string]string{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Println(err)
		helpers.WriteServerError(w, nil)
		return
	}
	newName := strings.TrimSpace(data["full_name"])
	if newName == "" {
		helpers.WriteError(w, http.StatusBadRequest, "username can't be empty", nil)
		return
	}
	err = uh.UserService.UpdateUserFullName(r.Context(), newName, userID)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteSuccess(w, http.StatusOK, "Name updated successfully", nil)
}
