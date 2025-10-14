package internal

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/set-kaung/senior_project_1/internal/helpers"
)

type ctxKey string

type statusRecorder struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	rec.body.Write(b)
	return rec.ResponseWriter.Write(b)
}

const UserIDContextKey ctxKey = "authenticatedUserID"

func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
			body:           &bytes.Buffer{},
		}
		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		if rec.status >= 500 {
			bodySnippet := rec.body.String()
			if len(bodySnippet) > 200 {
				bodySnippet = bodySnippet[:200] + "..."
			}
			loc := time.FixedZone("UTC+7", 7*60*60)
			now := time.Now().UTC()

			localTime := now.In(loc)
			helpers.WriteToWebHook(
				fmt.Sprintf("%s %s %s -> %d (%v) | Response: %s", localTime, r.Method, r.URL.Path, rec.status, duration, bodySnippet),
				os.Getenv("WEBHOOK_URL"),
			)
		}

	})
}

func getClerkUserID(ctx context.Context) (string, error) {
	claims, ok := clerk.SessionClaimsFromContext(ctx)
	if !ok {
		return "", ErrUnauthorized
	}
	clerkUser, err := user.Get(ctx, claims.Subject)
	if err != nil {
		return "", ErrInternalServerError
	}
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
	allowedOrigins := strings.Split(os.Getenv("REMOTE_ORIGINS"), ",")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if the request Origin is in the allowed list
		for _, o := range allowedOrigins {
			if strings.TrimSpace(o) == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin") // avoid caching issues
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, PATCH, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (i *SimpleRateLimiter) RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIP(r)
		limiter := i.GetLimiter(ip)

		if !limiter.Allow() {
			helpers.WriteError(w, http.StatusTooManyRequests, "Rate limit exceeded", nil)
			return
		}

		next.ServeHTTP(w, r)
	}
}
