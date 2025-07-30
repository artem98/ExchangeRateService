package rates

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Job struct {
	Currency1 string
	Currency2 string
	reqId     uint64
}

const queueSize = 200

var Jobs = make(chan Job, queueSize)
var hasStarted = false

func PlanJob(job Job) {
	if !hasStarted {
		start()
	}
	Jobs <- job
}

func start() {
	hasStarted = true
	go func() {
		for job := range Jobs {
			fmt.Println("Worker received currency pair", job)

			err := processJob(job)
			if err != nil {
				fmt.Println("Failed to process request ", job.reqId, ":", err)
			}
		}
	}()
}

const useRealExternalApi = false

func processJob(job Job) error {

	var rate float64
	var err error
	if useRealExternalApi {
		rate, err = fetchRateReal(job)
	} else {
		rate, err = fetchRateFake(job)
	}

	if err != nil {
		markRequestAsFailed(job.reqId)
		return err
	}

	err = updateRate(job.Currency1, job.Currency2, rate)
	if err != nil {
		markRequestAsFailed(job.reqId)
		return err
	}

	return markRequestAsProcessed(job.reqId)
}

var fakeRates = [...]float64{0.04, 0.89, 1.35, 33, 18.1, 9, 0.81, 0.33, 12.34, 2.93, 2.02, 1.09, 3.65, 0.11}
var it = 0

func fetchRateFake(job Job) (float64, error) {
	time.Sleep(3 * time.Second)
	it++
	it = it % len(fakeRates)
	return fakeRates[it], nil
}

const apiKey = "0cd612177560a71ffc4117930b976bb8"

type externalRateResponse struct {
	Rates map[string]float64 `json:"rates"`
	Base  string             `json:"base"`
}

func fetchRateReal(job Job) (float64, error) {
	url := fmt.Sprintf("https://api.exchangeratesapi.io/v1/latest?base=%s&symbols=%s&access_key=%s",
		strings.ToUpper(job.Currency1), strings.ToUpper(job.Currency2), apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API error: %s", resp.Status)
	}

	var parsed externalRateResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return 0, err
	}

	rate, ok := parsed.Rates[job.Currency2]
	if !ok {
		return 0, fmt.Errorf("rate for %s not found", job.Currency2)
	}

	return rate, nil
}
