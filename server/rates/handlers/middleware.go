package handlers

import (
	"fmt"
	"net/http"
)

type middleware = func(h http.HandlerFunc) http.HandlerFunc

func withRecovery(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Println("Recovered in handler:", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		h(w, r)
	}
}

func withLog(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handling", r.Method, " on url: ", r.URL)
		h(w, r)
	}
}

func handlerWithMiddleware(h http.HandlerFunc) http.HandlerFunc {
	mws := [...]middleware{withLog, withRecovery}

	for _, mw := range mws {
		h = mw(h)
	}

	return h
}
