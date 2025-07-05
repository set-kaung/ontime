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
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)

	listingRequest.Provider = user.User{ID: userID}
	_, err = lh.ListingService.CreateListing(r.Context(), listingRequest)
	if err != nil {
		log.Println("listing_handler -> HandleViewOwnProfile: ", err)
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
	id, _ := r.Context().Value(internal.UserIDContextKey).(string)

	listings, err := lh.ListingService.GetAllListings(r.Context(), id)
	if err != nil {
		log.Println("listing_handler -> HandleViewOwnProfile: ", err)
		helpers.WriteError(w, http.StatusInternalServerError, "user not found", nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, listings, nil)
}

func (lh *ListingHandler) HandleGetOwnListings(w http.ResponseWriter, r *http.Request) {
	id, _ := r.Context().Value(internal.UserIDContextKey).(string)

	listings, err := lh.ListingService.GetListingsByUserID(r.Context(), id)
	if err != nil {
		log.Println("listing_handler -> HandleViewOwnProfile: ", err)
		helpers.WriteError(w, http.StatusInternalServerError, "user not found", nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, listings, nil)
}

func (lh *ListingHandler) HandleUpdateListing(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	id, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		log.Printf("listing_handler -> HandleUpdateListing: failed to parse integer %v\n", err)
		helpers.WriteError(w, http.StatusBadRequest, "unprocessible entity", nil)
		return
	}
	decoder := json.NewDecoder(r.Body)
	listingRequest := Listing{}
	err = decoder.Decode(&listingRequest)
	if err != nil {
		log.Printf("listing_handler -> HandleUpdateListing: failed to decode json: %v\n", err)
		helpers.WriteError(w, http.StatusBadRequest, "bad request", nil)
		return
	}
	listingRequest.Provider.ID = userID
	listingRequest.ID = int32(id)
	lid, err := lh.ListingService.UpdateListing(r.Context(), listingRequest)
	if err != nil {
		log.Printf("listing_handler -> HandleUpdateListing: failed to update listing: %v\n", err)
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, map[string]int32{"id": lid}, nil)
}
