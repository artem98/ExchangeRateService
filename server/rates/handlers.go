package rates

import (
	// "encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func HandleRatesRequest(r chi.Router) {
	r.Route("/update_requests", func(r chi.Router) {
		r.Get("/", HandleGetRateByUpdateId)
		r.Post("/", HandlePostRateUpdateRequest)
		r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Only POST is allowed", http.StatusMethodNotAllowed)
		}))
	})
	r.Get("/", HandleGetRateByCode)
	r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Only GET is allowed", http.StatusMethodNotAllowed)
	}))
}

func HandleGetRateByCode(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received GET req by code")
}

func HandleGetRateByUpdateId(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received GET req by id")
}

func HandlePostRateUpdateRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received POST update req")
}
