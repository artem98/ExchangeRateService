package rates

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type UpdateResponse struct {
	UpdateID string `json:"update_request_id"`
}

func HandleRatesRequest(r chi.Router) {
	r.Route("/update_requests", func(r chi.Router) {
		r.Get("/{id}", HandleGetRateByUpdateId)
		r.Post("/", HandlePostRateUpdateRequest)
		r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Only POST and GET are allowed", http.StatusMethodNotAllowed)
		}))
	})
	r.Get("/", HandleGetRateByCode)
	r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
	}))
}

func HandleGetRateByCode(w http.ResponseWriter, r *http.Request) {
	currencyPair := r.URL.Query().Get("currency_pair")
	if currencyPair == "" {
		http.Error(w, "currency_pair query parameter is required", http.StatusBadRequest)
		return
	}

	fmt.Println("Recieved GET req by code", currencyPair)
}

func HandleGetRateByUpdateId(w http.ResponseWriter, r *http.Request) {
	updateId := chi.URLParam(r, "id")
	if updateId == "" {
		http.Error(w, "Update request id is required", http.StatusBadRequest)
		return
	}

	fmt.Println("Recieved GET req by id", updateId)
}

func HandlePostRateUpdateRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Recieved post req")

	resp := UpdateResponse{UpdateID: "blum"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
