package user

import (
	"context"
	"log"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/set-kaung/senior_project_1/internal"
)

func GetClerkUserID(ctx context.Context) (string, error) {
	claims, ok := clerk.SessionClaimsFromContext(ctx)
	if !ok {
		return "", internal.ErrUnauthorized
	}
	clerkUser, err := user.Get(ctx, claims.Subject)
	if err != nil {
		log.Printf("GetClerkUserID: failed to get user: %v", err)
		return "", internal.ErrInternalServerError
	}
	return clerkUser.ID, nil
}
