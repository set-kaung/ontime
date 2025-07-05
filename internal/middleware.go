package internal

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/set-kaung/senior_project_1/internal/helpers"
)

type ctxKey string

const UserIDContextKey ctxKey = "authenticatedUserID"

var logger *slog.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func LogMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request Started")
		next.ServeHTTP(w, r)
	})
}

func getClerkUserID(ctx context.Context) (string, error) {
	claims, ok := clerk.SessionClaimsFromContext(ctx)
	if !ok {
		return "", ErrUnauthorized
	}
	clerkUser, err := user.Get(ctx, claims.Subject)
	if err != nil {
		log.Printf("GetClerkUserID: failed to get user: %v", err)
		return "", ErrInternalServerError
	}
	log.Printf("Request from User %s with ID %s", *clerkUser.FirstName+*clerkUser.LastName, clerkUser.ID)
	return clerkUser.ID, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := getClerkUserID(r.Context())
		if err != nil {
			if err == ErrUnauthorized {
				helpers.WriteError(w, http.StatusUnauthorized, "unauthorized", nil)
			} else {
				helpers.WriteServerError(w, nil)
			}
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", os.Getenv("REMOTE_ORIGIN"))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
