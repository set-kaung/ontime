package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	"github.com/set-kaung/senior_project_1/internal/handlers"
	"github.com/set-kaung/senior_project_1/internal/repository"
	"github.com/set-kaung/senior_project_1/internal/service"
)

type application struct {
	userHandler handlers.UserHandler
}

func main() {

	port := flag.String("port", ":4096", "port to run the server. default is 4096. format - \":8080\"")

	flag.Parse()

	sqliteFile := "/Users/setkaung/sqlite_dbs/test_db.db"
	db, err := sql.Open("sqlite", sqliteFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalln("Ping failed:", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalln("Failed to enable foreign keys:", err)
	}

	sessionM := scs.New()
	sessionM.Store = sqlite3store.New(db)
	sessionM.Lifetime = 12 * time.Hour
	sessionM.Cookie.Secure = true

	a := &application{}

	userRepo := repository.NewSQLiteUserRepository(db)
	userService := service.UserService{Repo: userRepo}
	authService := service.AuthenticationService{Repo: userRepo}

	a.userHandler = handlers.UserHandler{UserService: userService, AuthService: authService}
	a.userHandler.SessionManager = sessionM

	mux := a.routes()
	log.Printf("starting server on port %s", *port)
	if err := http.ListenAndServe(*port, mux); err != nil {
		log.Fatalln(err)
	}
}
