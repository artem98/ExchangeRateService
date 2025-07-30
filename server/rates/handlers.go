package rates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type UpdateRequest struct {
	CurrencyPairCode string `json:"pair"`
}

type UpdateResponse struct {
	UpdateID string `json:"update_request_id"`
}

type RateResponse struct {
	Rate float64 `json:"rate"`
}

func HandleRatesRequest(r chi.Router) {
	r.Route("/update_requests", func(r chi.Router) {
		r.Get("/{id}", handleGetRateByUpdateId)
		r.Post("/", handlePostRateUpdateRequest)
		r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Only POST and GET are allowed", http.StatusMethodNotAllowed)
		}))
	})
	r.Get("/", handleGetRateByCode)
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

	fmt.Println("Recieved GET req by code", currencyPair)

	rate, err := getRateByPair(currencyPair)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RateResponse{Rate: rate})
}

func handleGetRateByUpdateId(w http.ResponseWriter, r *http.Request) {
	updateId := chi.URLParam(r, "id")
	if updateId == "" {
		http.Error(w, "Update request id is required", http.StatusBadRequest)
		return
	}

	fmt.Println("Recieved GET req by id", updateId)

	id, err := strconv.ParseUint(updateId, 10, 64)

	if err != nil {
		http.Error(w, "Update request id must be uint64", http.StatusBadRequest)
		return
	}

	rate, err := getRateByRequestId(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RateResponse{Rate: rate})
}

func handlePostRateUpdateRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Recieved post req_")

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

	fmt.Printf("  request %v\n", updateRequest)

	response := UpdateResponse{UpdateID: updateRequest.CurrencyPairCode}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
