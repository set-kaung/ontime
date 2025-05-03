package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/set-kaung/senior_project_1/internal/domain/user"
	"github.com/set-kaung/senior_project_1/internal/repository"
)

type application struct {
	userHandler user.UserHandler
}

func main() {

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

	sessionDB, err := sql.Open("postgres", os.Getenv("DBURL"))
	if err != nil {
		log.Fatalln(err)
	}
	defer sessionDB.Close()
	sessionM := scs.New()
	sessionM.Store = postgresstore.New(sessionDB)
	sessionM.Lifetime = 12 * time.Hour
	sessionM.Cookie.SameSite = http.SameSiteStrictMode
	sessionM.Cookie.Secure = true
	sessionM.Cookie.Persist = true  // Persist cookies across browser restarts
	sessionM.Cookie.HttpOnly = true // Recommended for security

	a := &application{}

	repo := repository.New(dbpool)

	userService := user.UserService{Repo: repo}

	a.userHandler = user.UserHandler{UserService: userService}
	a.userHandler.SessionManager = sessionM

	mux := a.routes()
	log.Printf("starting server on port %s", *port)
	if err := http.ListenAndServe(*port, mux); err != nil {
		log.Fatalln(err)
	}
}
