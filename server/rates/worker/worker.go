package worker

import (
	"fmt"

	"github.com/artem98/ExchangeRateService/server/rates/db"
	"github.com/artem98/ExchangeRateService/server/rates/external"
)

type Job struct {
	Currency1 string
	Currency2 string
	ReqId     uint64
}

const queueSize = 200

var jobs = make(chan Job, queueSize)

var hasStarted bool = false

func PlanJob(job Job) {
	if !hasStarted {
		start()
	}
	jobs <- job
}

func start() {
	hasStarted = true
	go func() {
		for job := range jobs {
			fmt.Println("Worker received currency pair", job)

			err := processJob(job)
			if err != nil {
				fmt.Println("Failed to process request ", job.ReqId, ":", err)
			}
		}
	}()
}

func processJob(job Job) (err error) {
	var rate float64
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in processJob:", r)
			db.MarkRequestAsFailed(job.ReqId)
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()

	rate, err = external.FetchRate(job.Currency1, job.Currency2)

	if err != nil {
		db.MarkRequestAsFailed(job.ReqId)
		return err
	}

	err = db.UpdateRate(job.Currency1, job.Currency2, rate)
	if err != nil {
		db.MarkRequestAsFailed(job.ReqId)
		return err
	}

	return db.MarkRequestAsProcessed(job.ReqId)
}
