package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/artem98/ExchangeRateService/server/rates/worker"
	"github.com/go-chi/chi/v5"
)

type mockDb struct {
	getByPair              func(cur1, cur2 string) (float64, time.Time, error)
	getByRequestId         func(id uint64) (float64, time.Time, error)
	placeRequest           func(cur1, cur2 string) (uint64, error)
	markRequestAsProcessed func(requestId uint64) error
	markRequestAsFailed    func(requestId uint64) error
	updateRate             func(currency1, currency2 string, rate float64) error
}

func (m *mockDb) GetRateByPair(currency1, currency2 string) (float64, time.Time, error) {
	return m.getByPair(currency1, currency2)
}
func (m *mockDb) GetRateByRequestId(id uint64) (float64, time.Time, error) {
	return m.getByRequestId(id)
}
func (m *mockDb) PlaceRequest(currency1, currency2 string) (uint64, error) {
	return m.placeRequest(currency1, currency2)
}
func (m *mockDb) MarkRequestAsProcessed(requestId uint64) error {
	return m.markRequestAsProcessed(requestId)
}
func (m *mockDb) MarkRequestAsFailed(requestId uint64) error {
	return m.markRequestAsFailed(requestId)
}
func (m *mockDb) UpdateRate(currency1, currency2 string, rate float64) error {
	return m.updateRate(currency1, currency2, rate)
}

type mockWorker struct {
	planned []worker.Job
}

func (m *mockWorker) PlanJob(job worker.Job) {
	m.planned = append(m.planned, job)
}

type mockCache struct {
	get func(currency1, currency2 string) (uint64, bool)
	set func(currency1, currency2 string, id uint64)
}

func (m *mockCache) Get(currency1, currency2 string) (uint64, bool) {
	return m.get(currency1, currency2)
}

func (m *mockCache) Set(currency1, currency2 string, id uint64) {
	m.set(currency1, currency2, id)
}

func TestHandleGetRateByCode(t *testing.T) {
	handler := &Handler{
		Db: &mockDb{
			getByPair: func(cur1, cur2 string) (float64, time.Time, error) {
				if cur1 == "EUR" && cur2 == "USD" {
					return 1.23, time.Now(), nil
				}
				return 0, time.Time{}, errors.New("not found")
			},
		},
		Worker: &mockWorker{},
		Cache:  &mockCache{},
	}

	req := httptest.NewRequest(http.MethodGet, "/?currency_pair=EUR/USD", nil)
	w := httptest.NewRecorder()

	handler.handleGetRateByCode(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}

	var resp RateResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if resp.Rate != 1.23 {
		t.Errorf("expected rate 1.23, got %f", resp.Rate)
	}
}

func TestHandleGetRateByCodeDBError(t *testing.T) {
	handler := &Handler{
		Db: &mockDb{
			getByPair: func(cur1, cur2 string) (float64, time.Time, error) {
				return 0, time.Time{}, errors.New("not found")
			},
		},
		Worker: &mockWorker{},
		Cache:  &mockCache{},
	}

	req := httptest.NewRequest(http.MethodGet, "/?currency_pair=EUR/USD", nil)
	w := httptest.NewRecorder()

	handler.handleGetRateByCode(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", res.StatusCode)
	}
}

func TestHandleGetRateByCodeIncorrectCode(t *testing.T) {
	handler := &Handler{
		Db:     &mockDb{},
		Worker: &mockWorker{},
		Cache:  &mockCache{},
	}

	req := httptest.NewRequest(http.MethodGet, "/?currency_pair=EURUSD", nil)
	w := httptest.NewRecorder()

	handler.handleGetRateByCode(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandleGetRateByCodeNoParam(t *testing.T) {
	handler := &Handler{
		Db:     &mockDb{},
		Worker: &mockWorker{},
		Cache:  &mockCache{},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.handleGetRateByCode(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandleGetRateByUpdateId(t *testing.T) {
	handler := &Handler{
		Db: &mockDb{
			getByRequestId: func(id uint64) (float64, time.Time, error) {
				if id == 42 {
					return 1.5, time.Now(), nil
				}
				return 0, time.Time{}, errors.New("not found")
			},
		},
		Worker: &mockWorker{},
		Cache:  &mockCache{},
	}

	r := chi.NewRouter()
	r.Get("/update_requests/{id}", handler.handleGetRateByUpdateId)

	req := httptest.NewRequest(http.MethodGet, "/update_requests/42", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
}

func TestHandleGetRateByUpdateIdNotUint64(t *testing.T) {
	handler := &Handler{
		Db: &mockDb{
			getByRequestId: func(id uint64) (float64, time.Time, error) {
				if id == 42 {
					return 1.5, time.Now(), nil
				}
				return 0, time.Time{}, errors.New("not found")
			},
		},
		Worker: &mockWorker{},
		Cache:  &mockCache{},
	}

	r := chi.NewRouter()
	r.Get("/update_requests/{id}", handler.handleGetRateByUpdateId)

	req := httptest.NewRequest(http.MethodGet, "/update_requests/42a", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", res.StatusCode)
	}
}

func TestHandleGetRateByUpdateIdDBError(t *testing.T) {
	handler := &Handler{
		Db: &mockDb{
			getByRequestId: func(id uint64) (float64, time.Time, error) {
				return 0, time.Time{}, errors.New("not found")
			},
		},
		Worker: &mockWorker{},
		Cache:  &mockCache{},
	}

	r := chi.NewRouter()
	r.Get("/update_requests/{id}", handler.handleGetRateByUpdateId)

	req := httptest.NewRequest(http.MethodGet, "/update_requests/42", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", res.StatusCode)
	}
}

func TestHandleGetRateByUpdateIdNoId(t *testing.T) {
	handler := &Handler{
		Db: &mockDb{
			getByRequestId: func(id uint64) (float64, time.Time, error) {
				if id == 42 {
					return 1.5, time.Now(), nil
				}
				return 0, time.Time{}, errors.New("not found")
			},
		},
		Worker: &mockWorker{},
		Cache:  &mockCache{},
	}

	r := chi.NewRouter()
	r.Get("/update_requests/{id}", handler.handleGetRateByUpdateId)

	req := httptest.NewRequest(http.MethodGet, "/update_requests/", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", res.StatusCode)
	}
}

func TestHandlePostRateUpdateRequestFoundInCache(t *testing.T) {
	mockW := &mockWorker{}
	handler := &Handler{
		Db: &mockDb{
			placeRequest: func(cur1, cur2 string) (uint64, error) {
				if cur1 == "EUR" && cur2 == "USD" {
					return 777, nil
				}
				return 0, errors.New("invalid")
			},
		},
		Worker: mockW,
		Cache: &mockCache{
			get: func(currency1, currency2 string) (uint64, bool) {
				return 189, true
			},
			set: func(currency1, currency2 string, id uint64) {},
		},
	}

	body := []byte(`{"pair":"EUR/USD"}`)
	req := httptest.NewRequest(http.MethodPost, "/update_requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.handlePostRateUpdateRequest(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}

	var resp UpdateResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if resp.UpdateID != 189 {
		t.Errorf("expected ID 189, got %d", resp.UpdateID)
	}
	if len(mockW.planned) != 0 {
		t.Errorf("expected no job to be planned")
	}
}

func TestHandlePostRateUpdateRequest(t *testing.T) {
	mockW := &mockWorker{}
	handler := &Handler{
		Db: &mockDb{
			placeRequest: func(cur1, cur2 string) (uint64, error) {
				if cur1 == "EUR" && cur2 == "USD" {
					return 777, nil
				}
				return 0, errors.New("invalid")
			},
		},
		Worker: mockW,
		Cache: &mockCache{
			get: func(currency1, currency2 string) (uint64, bool) {
				return 0, false
			},
			set: func(currency1, currency2 string, id uint64) {},
		},
	}

	body := []byte(`{"pair":"EUR/USD"}`)
	req := httptest.NewRequest(http.MethodPost, "/update_requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.handlePostRateUpdateRequest(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}

	var resp UpdateResponse
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}
	if resp.UpdateID != 777 {
		t.Errorf("expected ID 777, got %d", resp.UpdateID)
	}
	if len(mockW.planned) != 1 {
		t.Errorf("expected job to be planned")
	}
}

func TestHandlePostRateUpdateRequestNotJson(t *testing.T) {
	mockW := &mockWorker{}
	handler := &Handler{
		Db: &mockDb{
			placeRequest: func(cur1, cur2 string) (uint64, error) {
				if cur1 == "EUR" && cur2 == "USD" {
					return 777, nil
				}
				return 0, errors.New("invalid")
			},
		},
		Worker: mockW,
		Cache:  &mockCache{},
	}

	body := []byte(`{"pair":"EUR/USD"}`)
	req := httptest.NewRequest(http.MethodPost, "/update_requests", bytes.NewReader(body))

	w := httptest.NewRecorder()
	handler.handlePostRateUpdateRequest(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", res.StatusCode)
	}
}

func TestHandlePostRateUpdateRequestWrongJson(t *testing.T) {
	mockW := &mockWorker{}
	handler := &Handler{
		Db: &mockDb{
			placeRequest: func(cur1, cur2 string) (uint64, error) {
				if cur1 == "EUR" && cur2 == "USD" {
					return 777, nil
				}
				return 0, errors.New("invalid")
			},
		},
		Worker: mockW,
		Cache:  &mockCache{},
	}

	body := []byte(`{"currs:"EUR/USD"}`)
	req := httptest.NewRequest(http.MethodPost, "/update_requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.handlePostRateUpdateRequest(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlePostRateUpdateRequestWrongJsonValues(t *testing.T) {
	mockW := &mockWorker{}
	handler := &Handler{
		Db: &mockDb{
			placeRequest: func(cur1, cur2 string) (uint64, error) {
				if cur1 == "EUR" && cur2 == "USD" {
					return 777, nil
				}
				return 0, errors.New("invalid")
			},
		},
		Worker: mockW,
		Cache:  &mockCache{},
	}

	body := []byte(`{"currs":"EUR/USD"}`)
	req := httptest.NewRequest(http.MethodPost, "/update_requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.handlePostRateUpdateRequest(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlePostRateUpdateRequestWrongPair(t *testing.T) {
	mockW := &mockWorker{}
	handler := &Handler{
		Db: &mockDb{
			placeRequest: func(cur1, cur2 string) (uint64, error) {
				if cur1 == "EUR" && cur2 == "USD" {
					return 777, nil
				}
				return 0, errors.New("invalid")
			},
		},
		Worker: mockW,
		Cache:  &mockCache{},
	}

	body := []byte(`{"pair":"EUR/fUSD"}`)
	req := httptest.NewRequest(http.MethodPost, "/update_requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.handlePostRateUpdateRequest(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", res.StatusCode)
	}
}

func TestHandlePostRateUpdateRequestDBError(t *testing.T) {
	mockW := &mockWorker{}
	handler := &Handler{
		Db: &mockDb{
			placeRequest: func(cur1, cur2 string) (uint64, error) {
				return 0, errors.New("invalid")
			},
		},
		Worker: mockW,
		Cache: &mockCache{
			get: func(currency1, currency2 string) (uint64, bool) {
				return 0, false
			},
			set: func(currency1, currency2 string, id uint64) {},
		},
	}

	body := []byte(`{"pair":"EUR/USD"}`)
	req := httptest.NewRequest(http.MethodPost, "/update_requests", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.handlePostRateUpdateRequest(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", res.StatusCode)
	}
}
