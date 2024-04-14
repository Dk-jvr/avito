package main

import (
	"avito/Controller"
	"avito/Database"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	database := Database.InitDataBase()
	defer database.Close()

	router := mux.NewRouter()
	router.HandleFunc("/user_banner", Controller.UserBanner)
	router.HandleFunc("/banner", Controller.BannerProcessing)
	router.HandleFunc("/banner/{id}", Controller.AdminBannerProcessing)
	loginHandler := Controller.UserAuthMiddleware(router)

	log.Fatal(http.ListenAndServe(":8080", loginHandler))
}
