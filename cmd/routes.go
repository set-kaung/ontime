package main

import (
	"net/http"

	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
	"github.com/set-kaung/senior_project_1/internal"
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
	mux := http.NewServeMux()

	chain := NewRouteChainer()

	mux.Handle("GET /health", chain.Chain(HealthCheck))

	protected := chain.Append(internal.LogMiddleWare, clerkhttp.WithHeaderAuthorization(), internal.AuthMiddleware)

	mux.Handle("GET /users/me", protected.Chain(a.userHandler.HandleViewOwnProfile))
	mux.Handle("GET /users/me/services", protected.Chain(a.listingHandler.HandleGetOwnListings))
	mux.Handle("POST /update-profile-metadata", protected.Chain(a.userHandler.HandleInsertUser))
	mux.Handle("POST /users/me/update", protected.Chain(a.userHandler.HandleUpdateUserProfile))
	mux.Handle("DELETE /users/me/delete", protected.Chain(a.userHandler.HandleDeleteUser))

	mux.Handle("GET /services", protected.Chain(a.listingHandler.HandleGetAllListings))
	mux.Handle("GET /services/{id}", protected.Chain(a.listingHandler.HandleGetListingByID))
	mux.Handle("POST /services/create", protected.Chain(a.listingHandler.HandleCreateListing))
	mux.Handle("PUT /services/update/{id}", protected.Chain(a.listingHandler.HandleUpdateListing))

	mux.Handle("POST /requests/create/{id}", protected.Chain(a.requestHandler.HandleCreateRequest))
	mux.Handle("POST /requests/accept/{id}", protected.Chain(a.requestHandler.HandleAcceptServiceRequest))
	mux.Handle("POST /requests/decline/{id}", protected.Chain(a.requestHandler.HandleAcceptServiceRequest))
	mux.Handle("POST /requests/complete/{id}", protected.Chain(a.requestHandler.HandleCompleteServiceRequest))
	mux.Handle("GET /requests/{id}", protected.Chain(a.requestHandler.HandleGetRequestByID))
	mux.Handle("GET /requests/incoming", protected.Chain(a.requestHandler.HandleGetAllIncomingRequest))

	return internal.CORS(mux)
}
