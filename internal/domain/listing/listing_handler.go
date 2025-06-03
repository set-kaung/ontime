package listing

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
	"github.com/set-kaung/senior_project_1/internal/helpers"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type ListingHandler struct {
	listingService *listingService
}

func NewListingHandler(db *pgxpool.Pool) *ListingHandler {
	return &ListingHandler{listingService: &listingService{db}}
}

func (lh *ListingHandler) HandleCreateListing(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	listingRequest := repository.CreateListingParams{}
	err := decoder.Decode(&listingRequest)
	if err != nil {
		helpers.WriteError(w, http.StatusBadRequest, internal.ErrInternalServerError.Error(), nil)
		return
	}
	id, err := user.GetClerkUserID(r.Context())
	if err != nil {
		log.Println("user_handler -> HandleViewOwnProfile: ", err)
		helpers.WriteError(w, http.StatusInternalServerError, "user not found", nil)
		return
	}
	listingRequest.PostedBy = id
	_, err = lh.listingService.CreateListing(r.Context(), listingRequest)
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

	listing, err := lh.listingService.GetListingByID(r.Context(), int32(id))
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
		helpers.WriteError(w, http.StatusInternalServerError, "user not found", nil)
		return
	}
	listings, err := lh.listingService.GetAllListings(r.Context(), id)
	if err != nil {
		log.Println("user_handler -> HandleViewOwnProfile: ", err)
		helpers.WriteError(w, http.StatusInternalServerError, "user not found", nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, listings, nil)
}
