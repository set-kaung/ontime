package main

import (
	"net/http"

	"github.com/set-kaung/senior_project_1/internal/handlers"
)

type RouteChainer struct {
	routes []func(http.Handler) http.Handler
}

func NewRouteChainer(initial ...func(http.Handler) http.Handler) *RouteChainer {
	return &RouteChainer{routes: initial}
}

func (r *RouteChainer) Chain(next http.HandlerFunc) http.Handler {
	h := http.Handler(next)
	for i := range r.routes {
		h = r.routes[len(r.routes)-1-i](h)
	}

	return h
}

func (r *RouteChainer) Append(appendingRoutes ...func(http.Handler) http.Handler) *RouteChainer {
	newRoutes := make([]func(http.Handler) http.Handler, len(r.routes)+len(appendingRoutes))
	copy(newRoutes, r.routes)
	newRoutes = append(newRoutes, appendingRoutes...)
	return &RouteChainer{newRoutes}
}

func (a *application) routes() http.Handler {
	mux := http.NewServeMux()
	chain := NewRouteChainer(a.userHandler.SessionManager.LoadAndSave, a.userHandler.Authenticate)

	mux.Handle("GET /health", chain.Chain(handlers.HealthCheck))
	mux.Handle("POST /user/signup", chain.Chain(a.userHandler.HandleSignUp))
	mux.Handle("POST /user/login", chain.Chain(a.userHandler.HandleLogin))

	protected := chain.Append(a.userHandler.RequireAuthentication)

	// not complete yet
	mux.Handle("GET /user/profile/self", protected.Chain(a.userHandler.ViewOwnProfile))
	// mux.Handle("GET")
	return CORS(mux)
}
