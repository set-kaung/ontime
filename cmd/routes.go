package main

import (
	"net/http"
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
	newRoutes := make([]func(http.Handler) http.Handler, 0, len(r.routes)+len(appendingRoutes))
	newRoutes = append(newRoutes, r.routes...)
	newRoutes = append(newRoutes, appendingRoutes...)
	return &RouteChainer{newRoutes}
}

func (a *application) routes() http.Handler {
	mux := http.NewServeMux()

	chain := NewRouteChainer(a.userHandler.SessionManager.LoadAndSave, LogMiddleWare, a.userHandler.Authenticate)

	mux.Handle("GET /health", chain.Chain(HealthCheck))
	mux.Handle("POST /user/signup", chain.Chain(a.userHandler.HandleSignUp))
	mux.Handle("POST /user/login", chain.Chain(a.userHandler.HandleLogin))

	protected := chain.Append(a.userHandler.RequireAuthentication)

	// not complete yet
	mux.Handle("GET /user/profile/self", protected.Chain(a.userHandler.ViewOwnProfile))
	mux.Handle("POST /user/logout", protected.Chain(a.userHandler.HandleLogout))
	return CORS(mux)
}
