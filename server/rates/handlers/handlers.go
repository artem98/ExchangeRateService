package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/artem98/ExchangeRateService/server/rates/db"
	"github.com/artem98/ExchangeRateService/server/rates/utils"
	"github.com/artem98/ExchangeRateService/server/rates/worker"
	"github.com/go-chi/chi/v5"
)

type UpdateRequest struct {
	CurrencyPairCode string `json:"pair"`
}

type UpdateResponse struct {
	UpdateID uint64 `json:"update_request_id"`
}

type RateResponse struct {
	Rate      float64   `json:"rate"`
	Timestamp time.Time `json:"update_time"`
}

func HandleRatesRequest(r chi.Router) {
	r.Route("/update_requests", func(r chi.Router) {
		r.Get("/{id}", handlerWithMiddleware(handleGetRateByUpdateId))
		r.Post("/", handlerWithMiddleware(handlePostRateUpdateRequest))
		r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Only POST and GET are allowed", http.StatusMethodNotAllowed)
		}))
	})
	r.Get("/", handlerWithMiddleware(handleGetRateByCode))
	r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
	}))
}

func handleGetRateByCode(w http.ResponseWriter, r *http.Request) {
	currencyPair := r.URL.Query().Get("currency_pair")
	if currencyPair == "" {
		http.Error(w, "currency_pair query parameter is required", http.StatusBadRequest)
		return
	}

	currency1, currency2, err := utils.ParseCurrencyPair(currencyPair)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	rate, timestamp, err := db.GetRateByPair(currency1, currency2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(RateResponse{Rate: rate, Timestamp: timestamp})
	if err != nil {
		http.Error(w, "internal json problem", http.StatusInternalServerError)
	}
}

func handleGetRateByUpdateId(w http.ResponseWriter, r *http.Request) {
	updateId := chi.URLParam(r, "id")
	if updateId == "" {
		http.Error(w, "Update request id is required", http.StatusNotFound)
		return
	}

	id, err := strconv.ParseUint(updateId, 10, 64)

	if err != nil {
		http.Error(w, "Update request id must be uint64", http.StatusNotFound)
		return
	}

	rate, timestamp, err := db.GetRateByRequestId(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(RateResponse{Rate: rate, Timestamp: timestamp})
	if err != nil {
		http.Error(w, "internal json problem", http.StatusInternalServerError)
	}
}

func handlePostRateUpdateRequest(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var updateRequest UpdateRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&updateRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if updateRequest.CurrencyPairCode == "" {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	fmt.Printf("  request: %v\n", updateRequest)

	currency1, currency2, err := utils.ParseCurrencyPair(updateRequest.CurrencyPairCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	requestId, err := db.PlaceRequest(currency1, currency2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	worker.PlanJob(worker.Job{Currency1: currency1, Currency2: currency2, ReqId: requestId})

	response := UpdateResponse{UpdateID: requestId}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "internal json problem", http.StatusInternalServerError)
	}
}
