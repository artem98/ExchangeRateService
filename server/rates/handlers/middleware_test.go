package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRecovery_NoPanic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	withRecovery(handler).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestWithRecovery_Panic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	withRecovery(handler).ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Internal Server Error") {
		t.Errorf("unexpected response body: %s", w.Body.String())
	}
}

func TestWithLog_CallsHandler(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	withLog(handler).ServeHTTP(w, r)

	if !called {
		t.Error("handler was not called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestHandleWithMiddlewaresCalled(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	handlerWithMiddleware(handler).ServeHTTP(w, r)

	if !called {
		t.Error("final handler was not called")
	}
}
