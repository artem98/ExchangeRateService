package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	// "encoding/json"

	"github.com/artem98/ExchangeRateService/server/rates/db"
	"github.com/artem98/ExchangeRateService/server/rates/handlers"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello client!")
}

func main() {
	err := db.InitDataBaseInterface()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.CloseDB()

	router := chi.NewRouter()
	router.Route("/rates", handlers.HandleRatesRequest)
	router.HandleFunc("/", defaultHandler)

	fmt.Println("Server started")
	http.ListenAndServe(":8080", router)
}
