package listing

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
	"github.com/set-kaung/senior_project_1/internal/helpers"
)

type ListingHandler struct {
	ListingService ListingService
}

func (lh *ListingHandler) HandleCreateListing(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	listingRequest := Listing{}
	err := decoder.Decode(&listingRequest)
	if err != nil {
		log.Println(err)
		helpers.WriteError(w, http.StatusBadRequest, "bad request", nil)
		return
	}
	id, err := user.GetClerkUserID(r.Context())
	if err != nil {
		log.Println("user_handler -> HandleViewOwnProfile: ", err)
		helpers.WriteError(w, http.StatusInternalServerError, "user not found", nil)
		return
	}
	listingRequest.Provider = user.User{ID: id}
	_, err = lh.ListingService.CreateListing(r.Context(), listingRequest)
	if err != nil {
		log.Println("user_handler -> HandleViewOwnProfile: ", err)
		helpers.WriteError(w, http.StatusInternalServerError, "error creating listing", nil)
		return
	}
}

func (lh *ListingHandler) HandleGetListingByID(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	id, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}

	listing, err := lh.ListingService.GetListingByID(r.Context(), int32(id))
	if err != nil {
		helpers.WriteError(w, http.StatusInternalServerError, internal.ErrInternalServerError.Error(), nil)
		return
	}
	err = helpers.WriteData(w, http.StatusOK, listing, nil)
	if err != nil {
		log.Println("listing_handler -> HandleGetListingByID: ", err)
	}
}

func (lh *ListingHandler) HandleGetAllListings(w http.ResponseWriter, r *http.Request) {
	id, err := user.GetClerkUserID(r.Context())
	if err != nil {
		log.Println("user_handler -> HandleViewOwnProfile: ", err)
		helpers.WriteError(w, http.StatusInternalServerError, "unauthorized", nil)
		return
	}
	listings, err := lh.ListingService.GetAllListings(r.Context(), id)
	if err != nil {
		log.Println("user_handler -> HandleViewOwnProfile: ", err)
		helpers.WriteError(w, http.StatusInternalServerError, "user not found", nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, listings, nil)
}
