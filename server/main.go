package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	// "encoding/json"

	"github.com/artem98/ExchangeRateService/server/rates"
)

var reqCount int = 0

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello client!", reqCount)
	reqCount++
}

func main() {
	err := rates.InitDataBaseInterface()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer rates.CloseDB()

	router := chi.NewRouter()

	router.Route("/rates", rates.HandleRatesRequest)

	router.HandleFunc("/", handler)

	fmt.Println("Server started")
	http.ListenAndServe(":8080", router)
}
