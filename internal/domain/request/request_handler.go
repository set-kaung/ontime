package request

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
	"github.com/set-kaung/senior_project_1/internal/helpers"
)

type RequestHandler struct {
	RequestService RequestService
}

func (rh *RequestHandler) HandleCreateRequest(w http.ResponseWriter, r *http.Request) {
	userID, err := user.GetClerkUserID(r.Context())
	if err != nil && errors.Is(err, internal.ErrUnauthorized) {
		helpers.WriteError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	} else if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	pathID := r.PathValue("id")
	listingID, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		log.Println("request_handler -> HandleCreateRequest: err: ", err)
		helpers.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}
	serviceRequest := CreateClientServiceRequest(int32(listingID), userID)
	id, err := rh.RequestService.CreateServiceRequest(r.Context(), serviceRequest)
	if err != nil {
		log.Println("request_handler -> HandleCreateRequest: err: ", err)
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, map[string]int32{"requestID": id}, nil)
}

func (rh *RequestHandler) HandleGetAllIncomingRequest(w http.ResponseWriter, r *http.Request) {
	userID, err := user.GetClerkUserID(r.Context())
	if err != nil && errors.Is(err, internal.ErrUnauthorized) {
		helpers.WriteError(w, http.StatusUnauthorized, err.Error(), nil)
		return
	} else if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	requests, err := rh.RequestService.GetAllIncomingRequests(r.Context(), userID)
	if err != nil {
		log.Println("request_handler -> HandleGetAllIncomingRequest: err: ", err)
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, requests, nil)

}
