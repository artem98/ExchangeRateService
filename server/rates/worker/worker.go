package worker

import (
	"fmt"

	"github.com/artem98/ExchangeRateService/server/constants"
)

type Job = func() error

type Worker struct {
	hasStarted bool
	jobs       chan Job
}

func MakeWorker() *Worker {
	return &Worker{hasStarted: false, jobs: make(chan Job, constants.WorkerQueueSize)}
}

func (w *Worker) PlanJob(job Job) {
	if !w.hasStarted {
		w.start()
	}
	w.jobs <- job
}

func (w *Worker) start() {
	w.hasStarted = true
	go func() {
		for job := range w.jobs {
			err := w.processJob(job)
			if err != nil {
				fmt.Printf("Failed to process job %v : %s", job, err.Error())
			}
		}
	}()
}

func (w *Worker) processJob(job Job) (err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in processJob:", r)
			err = fmt.Errorf("panic occurred: %v", r)
		}
	}()

	fmt.Println("Worker starts processing new job")

	return job()
}
