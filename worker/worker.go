package worker

import (
	"fmt"
	"log"
	"math"

	"time"

	"github.com/Avik-creator/queue"
	"github.com/Avik-creator/utils"
)

type Worker struct {
	ID    int
	Queue *queue.JobQueue
}

func (w *Worker) Start() {
	go func() {
		for {
			j, err := w.Queue.GetJob()
			if err != nil {
				fmt.Println("No job found, sleeping for 1 second")
				time.Sleep(1 * time.Second)
				continue
			}
			if j.ID != "" {
				fmt.Printf("Worker %d processing job ID : %s \n", w.ID, j.ID)

				err := handleJob(j)
				if err != nil {
					log.Printf("Job %s failed : %v\n", j.ID, err)

					j.RetryCount++
					if j.RetryCount <= j.MaxRetries {
						delay := time.Duration(math.Pow(2, float64(j.RetryCount))) * time.Second
						log.Printf("Retrying job %s in %v \n", j.ID, delay)

						go func(jobCopy utils.Job) {
							time.Sleep(delay)
							w.Queue.AddJob(jobCopy)
						}(j)
					} else {
						log.Printf("Job %s moved dto dead-letter queue \n", j.ID)
						w.Queue.MoveJobToDeadLetterQueue(j)
					}
				}
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}()
}

func handleJob(j utils.Job) error {
	time.Sleep(500 * time.Millisecond)

	if j.Payload["to"] == "error@error.com" {
		return fmt.Errorf("simulated error")
	}

	fmt.Printf("Handled job : %s for %s\n", j.ID, j.Payload["to"])
	return nil
}
