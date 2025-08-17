package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal"
	"github.com/set-kaung/senior_project_1/internal/domain/listing"

	"github.com/set-kaung/senior_project_1/internal/domain/request"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
)

type application struct {
	userHandler    *user.UserHandler
	listingHandler *listing.ListingHandler
	requestHandler *request.RequestHandler
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file: ", err)
		log.Println("Using system defaults.")
	}

	clerkKey := os.Getenv("CLERK_SECRET_KEY")

	if clerkKey == "" {
		log.Fatalln("can't load clerk key.")
		return
	}

	clerk.SetKey(clerkKey)
	initCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	dbURL := os.Getenv("DBURL")
	if dbURL == "" {
		log.Fatalln("can't load db url")
		return
	}

	dbpool, err := pgxpool.New(initCtx, dbURL)
	if err != nil {
		log.Fatalf("error creating a pgxpool: %v\n", err)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalln("ping failed:", err)
		return
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("invalid port: ", port)
		log.Println("Using default port 8080")
		port = "8080"
	}
	tokenReward := os.Getenv("ONETIME_PAYMENT_TOKENS")
	_, err = strconv.Atoi(tokenReward)
	if err != nil {
		panic(err)
	}

	internal.PusherClient = internal.NewPusherClient()

	a := &application{}

	psqlUserService := &user.PostgresUserService{DB: dbpool}
	psqlListingService := &listing.PostgresListingService{DB: dbpool}
	psqlRequestService := &request.PostgresRequestService{DB: dbpool}

	a.userHandler = &user.UserHandler{UserService: psqlUserService}
	a.listingHandler = &listing.ListingHandler{ListingService: psqlListingService}
	a.requestHandler = &request.RequestHandler{RequestService: psqlRequestService}

	mux := a.routes()

	log.Printf("starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalln(err)
	}
}
