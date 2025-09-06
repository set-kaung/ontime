package main

import (
	"net/http"
	"time"

	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/set-kaung/senior_project_1/internal"
	"golang.org/x/time/rate"
)

type RouteChainer struct {
	routes []func(http.Handler) http.Handler
}

func NewRouteChainer(initial ...func(http.Handler) http.Handler) *RouteChainer {
	return &RouteChainer{routes: initial}
}
func (r *RouteChainer) Chain(next http.HandlerFunc) http.Handler {
	h := http.Handler(next)
	for i := len(r.routes) - 1; i >= 0; i-- {
		h = r.routes[i](h)
	}
	return h
}

func (r *RouteChainer) Append(appendingRoutes ...func(http.Handler) http.Handler) *RouteChainer {
	newRoutes := make([]func(http.Handler) http.Handler, 0, len(r.routes)+len(appendingRoutes))
	newRoutes = append(newRoutes, r.routes...)
	newRoutes = append(newRoutes, appendingRoutes...)
	return &RouteChainer{newRoutes}
}

func (a *application) routes() http.Handler {

	limiter := internal.NewSimpleRateLimiter(rate.Every(time.Second*10), 20)
	mux := http.NewServeMux()

	chain := NewRouteChainer()

	mux.Handle("GET /health", chain.Chain(HealthCheck))

	protected := chain.Append(internal.LogMiddleWare, clerkhttp.WithHeaderAuthorization(), internal.AuthMiddleware)

	mux.Handle("GET /users/me", protected.Chain(a.userHandler.HandleViewOwnProfile))
	mux.Handle("GET /users/me/services", protected.Chain(a.listingHandler.HandleGetOwnListings))
	mux.Handle("POST /update-profile-metadata", protected.Chain(a.userHandler.HandleInsertUser))
	mux.Handle("POST /users/me/update", protected.Chain(a.userHandler.HandleUpdateUserProfile))
	mux.Handle("PATCH /users/me/change-name", protected.Chain(a.userHandler.HandleUpdateUserFullName))
	mux.Handle("DELETE /users/me/delete", protected.Chain(a.userHandler.HandleDeleteUser))
	mux.Handle("GET /notifications", protected.Chain(limiter.RateLimitMiddleware(a.userHandler.GetUserNotifications)))
	mux.Handle("PUT /read-notification", protected.Chain(a.userHandler.HandleUpdateNotificationStatus))
	mux.Handle("GET /users/me/rewards", protected.Chain(a.rewardHandler.HandleGetAllUserRedeemdRewards))
	mux.Handle("GET /users/me/history", protected.Chain(a.userHandler.HandleGetAllHistories))
	mux.Handle("GET /users/me/redeemed-rewards/{redemptionId}", protected.Chain(a.rewardHandler.HandleGetRedeemedRewardByID))
	mux.Handle("PUT /notifications/mark-all-read", protected.Chain(a.userHandler.HandleMarkAllAsRead))
	mux.Handle("GET /users/me/completed-transactions/{requestId}", protected.Chain(a.requestHandler.GetCompletedTransaction))

	mux.Handle("GET /services", protected.Chain(a.listingHandler.HandleGetAllListings))
	mux.Handle("GET /services/{id}", protected.Chain(a.listingHandler.HandleGetListingByID))
	mux.Handle("POST /services/create", protected.Chain(a.listingHandler.HandleCreateListing))
	mux.Handle("PUT /services/update/{id}", protected.Chain(a.listingHandler.HandleUpdateListing))
	mux.Handle("POST /services/report/{id}", protected.Chain(a.listingHandler.HandleReportListing))
	mux.Handle("DELETE /services/delete/{id}", protected.Chain(a.listingHandler.HandleDeleteListing))
	mux.Handle("GET /services/{id}/reviews", protected.Chain(a.listingHandler.HandleGetListingReviews))

	mux.Handle("POST /requests/create/{id}", protected.Chain(a.requestHandler.HandleCreateRequest))
	mux.Handle("POST /requests/accept/{id}", protected.Chain(a.requestHandler.HandleAcceptServiceRequest))
	mux.Handle("POST /requests/decline/{id}", protected.Chain(a.requestHandler.HandleDeclineServiceRequest))
	mux.Handle("POST /requests/complete/{id}", protected.Chain(a.requestHandler.HandleCompleteServiceRequest))
	mux.Handle("GET /requests/{id}", protected.Chain(a.requestHandler.HandleGetRequestByID))
	mux.Handle("GET /requests/all", protected.Chain(a.requestHandler.HandleGetAllUserRequests))
	mux.Handle("POST /requests/review/{id}", protected.Chain(a.reviewHandler.HandleSubmitReview))

	mux.Handle("POST /ads/complete", protected.Chain(a.userHandler.HandleAdWatched))
	mux.Handle("GET /ads/watched", protected.Chain(a.userHandler.HandleGetAdsWatched))

	mux.Handle("GET /rewards", protected.Chain(a.rewardHandler.HandleGetAllRewards))
	mux.Handle("GET /rewards/{id}", protected.Chain(a.rewardHandler.HandleRewardByID))
	mux.Handle("POST /rewards/redeem/{id}", protected.Chain(a.rewardHandler.HandleRedeemReward))

	mux.Handle("GET /reviews/{id}", protected.Chain(a.reviewHandler.HandleGetReviewByID))
	return internal.CORS(mux)
}
