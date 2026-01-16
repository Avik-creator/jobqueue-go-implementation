package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Avik-creator/queue"
	"github.com/Avik-creator/scheduler"
	"github.com/Avik-creator/utils"
	"github.com/Avik-creator/worker"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

var q = queue.NewQueue()
var s = scheduler.NewScheduler(q)

func main() {
	StartCLI()
}

func StartCLI() {
	app := &cli.App{
		Name:  "Job Queue CLI",
		Usage: "Manage jobs, workers, and queues",
		Commands: []*cli.Command{
			{
				Name:  "enqueue",
				Usage: "Enqueue a job",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "to", Required: true},
					&cli.StringFlag{Name: "priority", Value: "low"},
					&cli.IntFlag{Name: "retries", Value: 3},
					&cli.IntFlag{Name: "delay", Value: 0, Usage: "Delay in seconds"},
				},
				Action: func(c *cli.Context) error {
					priority := utils.Low
					if c.String("priority") == "high" {
						priority = utils.High
					}

					j := utils.Job{
						ID:         uuid.New().String(),
						Type:       "email",
						Payload:    map[string]string{"to": c.String("to")},
						Priority:   priority,
						MaxRetries: c.Int("retries"),
						CreatedAt:  time.Now(),
					}

					delay := c.Int("delay")
					if delay > 0 {
						s.Scheduler(j, time.Duration(delay)*time.Second)
						fmt.Println("Scheduled job:", j.ID)
					} else {
						q.AddJob(j)
						fmt.Println("Enqueued job:", j.ID)
					}
					return nil
				},
			},
			{
				Name:  "start",
				Usage: "Start worker(s)",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "count", Value: 1},
				},
				Action: func(c *cli.Context) error {
					count := c.Int("count")
					for i := 1; i <= count; i++ {
						w := &worker.Worker{ID: i, Queue: q}
						w.Start()
					}
					fmt.Printf("Started %d worker(s)\n", count)
					select {} // block forever
				},
			},
			{
				Name:  "dlq",
				Usage: "Show dead-letter queue",
				Action: func(c *cli.Context) error {
					highJobs, mediumJobs, lowJobs, err := q.GetAllDeadLetterJobs()
					if err != nil {
						return fmt.Errorf("failed to get dead letter jobs: %v", err)
					}

					allJobs := append(append(highJobs, mediumJobs...), lowJobs...)

					if len(allJobs) == 0 {
						fmt.Println("No failed jobs")
						return nil
					}
					fmt.Println("Dead-letter jobs:")
					for _, j := range allJobs {
						fmt.Printf("- %s (%s, priority: %v, retries: %d/%d)\n", j.ID, j.Payload["to"], j.Priority, j.RetryCount, j.MaxRetries)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println("Error:", err)
	}
}
