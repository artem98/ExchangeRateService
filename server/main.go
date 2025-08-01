package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	// "encoding/json"

	"github.com/artem98/ExchangeRateService/server/rates/db"
	"github.com/artem98/ExchangeRateService/server/rates/handlers"
	"github.com/artem98/ExchangeRateService/server/rates/worker"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello client!")
}

func main() {
	dbAdapter, err := db.MakeDataBaseAdapter()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer dbAdapter.CloseDB()

	ratesHandler := &handlers.Handler{
		Db:     dbAdapter,
		Worker: worker.MakeWorker(),
	}

	router := chi.NewRouter()
	router.Route("/rates", ratesHandler.HandleRates)
	router.HandleFunc("/", defaultHandler)

	fmt.Println("Server started")
	http.ListenAndServe(":8080", router)
}
