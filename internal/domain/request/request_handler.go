package request

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/helpers"
)

type RequestHandler struct {
	RequestService RequestService
}

func (rh *RequestHandler) HandleCreateRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	pathID := r.PathValue("id")
	listingID, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		log.Println("request_handler -> HandleCreateRequest: err: ", err)
		helpers.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}
	serviceRequest := CreateClientServiceRequest(int32(listingID), userID)
	requestID, err := rh.RequestService.CreateServiceRequest(r.Context(), serviceRequest)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Println("request_handler -> HandleCreateRequest: err: ", err)
			helpers.WriteError(w, http.StatusBadRequest, "invalid request", nil)
			return
		}
		if errors.Is(err, internal.ErrInsufficientBalance) {
			helpers.WriteError(w, http.StatusBadRequest, "insufficient balance", nil)
			return
		}
		log.Println("request_handler -> HandleCreateRequest: err: ", err)
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusCreated, map[string]int32{"requestID": requestID}, nil)
}

func (rh *RequestHandler) HandleGetRequestByID(w http.ResponseWriter, r *http.Request) {
	pathID := r.PathValue("id")
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	requestID, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		log.Println("request_handler -> HandleGetRequestByID: err: ", err)
		helpers.WriteError(w, http.StatusBadRequest, "unprocessable entity", nil)
		return
	}
	request, err := rh.RequestService.GetRequestByID(r.Context(), int32(requestID))
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	request.IsProvider = userID == request.Provider.ID
	if request.Provider.ID == userID {
		request.Type = "INCOMING"
	} else {
		request.Type = "OUTGOING"
	}

	fmt.Println(request)
	helpers.WriteData(w, http.StatusOK, request, nil)
}

func (rh *RequestHandler) HandleGetAllUserRequests(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	requests, err := rh.RequestService.GetUserActiveServiceRequests(r.Context(), userID)

	if err != nil {
		log.Println("request_handler -> HandleGetAllIncomingRequest: err: ", err)
		helpers.WriteServerError(w, nil)
		return
	}

	helpers.WriteData(w, http.StatusOK, requests, nil)

}

func (rh *RequestHandler) HandleAcceptServiceRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	pathID := r.PathValue("id")
	listingID, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		log.Println("request_handler -> HandleAcceptServiceRequest: err: ", err)
		helpers.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}
	rid, err := rh.RequestService.AcceptServiceRequest(r.Context(), int32(listingID), userID)
	if err != nil {
		log.Println("request_handler -> HandleAcceptServiceRequest: err: ", err)
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, map[string]int32{"request_id": rid}, nil)
}

func (rh *RequestHandler) HandleDeclineServiceRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	pathID := r.PathValue("id")
	listingID, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		log.Println("request_handler -> HandleAcceptServiceRequest: err: ", err)
		helpers.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}
	rid, err := rh.RequestService.DeclineServiceRequest(r.Context(), int32(listingID), userID)
	if err != nil {
		log.Println("request_handler -> HandleAcceptServiceRequest: err: ", err)
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, map[string]int32{"request_id": rid}, nil)
}

func (rh *RequestHandler) HandleCompleteServiceRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	pathID := r.PathValue("id")
	requestID, err := strconv.ParseInt(pathID, 10, 32)
	if err != nil {
		log.Println("request_handler -> HandleAcceptServiceRequest: err: ", err)
		helpers.WriteError(w, http.StatusBadRequest, "invalid id", nil)
		return
	}
	rid, err := rh.RequestService.CompleteServiceRequest(r.Context(), int32(requestID), userID)
	if err != nil {
		if errors.Is(err, internal.ErrUnauthorized) {
			helpers.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)

		} else {
			helpers.WriteServerError(w, nil)
		}
		return
	}

	helpers.WriteData(w, http.StatusOK, map[string]int32{"request_id": rid}, nil)
}

func (rh *RequestHandler) HandleGetCompletedTransaction(w http.ResponseWriter, r *http.Request) {

	requestPathValue := r.PathValue("requestId")
	requestID, err := strconv.ParseInt(requestPathValue, 10, 32)
	if err != nil {
		log.Printf("GetCompletedTransaction: %s \n", err)
		helpers.WriteError(w, http.StatusUnprocessableEntity, "unprocessable entity", nil)
		return
	}
	request, err := rh.RequestService.GetRequestByID(r.Context(), int32(requestID))
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, request, nil)
}

func (rh *RequestHandler) HandleCreateRequestReport(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	requestPathValue := r.PathValue("id")
	requestID, err := strconv.ParseInt(requestPathValue, 10, 32)
	if err != nil {
		log.Printf("HandleCreateRequestReport: %s \n", err)
		helpers.WriteError(w, http.StatusUnprocessableEntity, "unprocessable entity", nil)
		return
	}
	ticketID, err := rh.RequestService.CreateRequestReport(r.Context(), int32(requestID), userID)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusCreated, map[string]string{"ticket_id": ticketID}, nil)
}

func (rh *RequestHandler) HandleGetReviewByRequestID(w http.ResponseWriter, r *http.Request) {
	requestPathValue := r.PathValue("id")
	requestID, err := strconv.ParseInt(requestPathValue, 10, 32)
	if err != nil {
		log.Printf("HandleGetReviewByRequestID: %s \n", err)
		helpers.WriteError(w, http.StatusUnprocessableEntity, "unprocessable entity", nil)
		return
	}
	review, err := rh.RequestService.GetRequestReview(r.Context(), int32(requestID))
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, review, nil)
}

func (rh *RequestHandler) HandleGetRequestReport(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	requestPathValue := r.PathValue("id")
	requestID, err := strconv.ParseInt(requestPathValue, 10, 32)
	if err != nil {
		log.Printf("HandleGetRequestReport: %s \n", err)
		helpers.WriteError(w, http.StatusUnprocessableEntity, "unprocessable entity", nil)
		return
	}
	report, err := rh.RequestService.GetRequestReport(r.Context(), int32(requestID), userID)
	if err != nil {
		if errors.Is(err, internal.ErrNoRecord) {
			helpers.WriteError(w, http.StatusNotFound, "no such recor", nil)
			return
		}
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, report, nil)
}

func (rh *RequestHandler) HandleCancelRequest(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	requestPathValue := r.PathValue("id")
	requestID, err := strconv.ParseInt(requestPathValue, 10, 32)
	if err != nil {
		log.Printf("HandleCancelRequest: %s \n", err)
		helpers.WriteError(w, http.StatusUnprocessableEntity, "unprocessable entity", nil)
		return
	}
	err = rh.RequestService.CancelServiceRequest(r.Context(), int32(requestID), userID)
	if err != nil {
		if errors.Is(err, internal.ErrUnauthorized) {
			helpers.WriteError(w, http.StatusUnauthorized, "not authorized", nil)
		} else {
			helpers.WriteServerError(w, nil)
		}
		return
	}
	helpers.WriteSuccess(w, http.StatusOK, "request cancelled", nil)
}

func (rh *RequestHandler) HandleGetAllUserRequestReports(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(internal.UserIDContextKey).(string)
	tickets, err := rh.RequestService.GetAllUserRequestReports(r.Context(), userID)
	if err != nil {
		helpers.WriteServerError(w, nil)
		return
	}
	helpers.WriteData(w, http.StatusOK, tickets, nil)
}
