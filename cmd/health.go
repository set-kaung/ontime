package main

import (
	"log"
	"net/http"

	"github.com/set-kaung/senior_project_1/internal/helpers"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := helpers.WriteSuccess(w, 200, "Up and Running!", nil); err != nil {
		log.Println(err)
	}
}
