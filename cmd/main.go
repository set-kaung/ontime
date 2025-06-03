package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal/domain/listing"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
)

type application struct {
	userHandler    *user.UserHandler
	listingHandler *listing.ListingHandler
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clerk.SetKey(os.Getenv("CLARKE_SECRET_KEY"))

	fmt.Printf("accepting request from %s \n", os.Getenv("REMOTE_ORIGIN"))

	port := flag.String("port", ":4096", "port to run the server. default is 4096. format - \":8080\"")

	flag.Parse()

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DBURL"))
	if err != nil {
		log.Fatalln(err)
	}
	defer dbpool.Close()
	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalln("Ping failed:", err)
	}

	// sessionDB, err := sql.Open("postgres", os.Getenv("DBURL"))
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer sessionDB.Close()
	// sessionM := scs.New()
	// // sessionM.Store = postgresstore.New(sessionDB)
	// sessionM.Lifetime = 12 * time.Hour
	// sessionM.Cookie.SameSite = http.SameSiteNoneMode
	// sessionM.Cookie.Secure = true
	// sessionM.Cookie.Persist = true  // Persist cookies across browser restarts
	// sessionM.Cookie.HttpOnly = true // Recommended for security

	a := &application{}

	a.userHandler = user.NewUserHandler(dbpool)
	a.listingHandler = listing.NewListingHandler(dbpool)

	mux := a.routes()
	log.Printf("starting server on port %s", *port)
	if err := http.ListenAndServe(*port, mux); err != nil {
		log.Fatalln(err)
	}
}
