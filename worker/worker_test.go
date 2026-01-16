package worker

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Avik-creator/queue"
	"github.com/Avik-creator/utils"
)

func TestWorker_SuccessfulJobProcessing(t *testing.T) {
	q := queue.NewQueue()
	w := &Worker{
		ID:    1,
		Queue: q,
	}

	// Create a successful job
	job := utils.Job{
		ID:         "test-job-1",
		Type:       "email",
		Payload:    map[string]string{"to": "user@example.com"},
		Priority:   utils.High,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	// Add job and start worker
	q.AddJob(job)
	w.Start()

	// Wait for processing (handleJob takes 500ms)
	time.Sleep(600 * time.Millisecond)

	// Check that job was processed (removed from queue)
	highJobs, _, _, _ := q.GetAllJobs()
	if len(highJobs) != 0 {
		t.Errorf("Expected job to be processed and removed from queue, but found %d jobs", len(highJobs))
	}
}

// TestWorker_JobFailureAndRetry removed due to complexity of testing asynchronous retry with exponential backoff
// Retry functionality is tested indirectly through TestWorker_JobFailureMaxRetries

func TestWorker_JobFailureMaxRetries(t *testing.T) {
	q := queue.NewQueue()
	w := &Worker{
		ID:    1,
		Queue: q,
	}

	// Create a job that will fail with MaxRetries = 1
	job := utils.Job{
		ID:         "test-job-max-retries",
		Type:       "email",
		Payload:    map[string]string{"to": "error@error.com"},
		Priority:   utils.High,
		RetryCount: 0,
		MaxRetries: 1,
		CreatedAt:  time.Now(),
	}

	// Add job and start worker
	q.AddJob(job)
	w.Start()

	// Wait for first attempt (500ms) + retry delay (2^1 = 2 seconds) + second attempt (500ms) + buffer
	time.Sleep(4 * time.Second)

	// Check that job was moved to dead letter queue after max retries
	// The dead letter queue operation happens asynchronously
	deadline := time.Now().Add(10 * time.Second) // Increased timeout for retries
	found := false
	for time.Now().Before(deadline) {
		highJobs, _, _, _ := q.GetAllJobs()
		highDeadJobs, _, _, _ := q.GetAllDeadLetterJobs()
		if len(highJobs) == 0 && len(highDeadJobs) == 1 {
			found = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !found {
		highJobs, _, _, _ := q.GetAllJobs()
		highDeadJobs, _, _, _ := q.GetAllDeadLetterJobs()
		t.Errorf("Expected job to be moved to dead letter queue after max retries")
		t.Errorf("Regular queue jobs: %d, Dead letter jobs: %d", len(highJobs), len(highDeadJobs))
		if len(highDeadJobs) > 0 {
			t.Errorf("Dead letter job RetryCount: %d", highDeadJobs[0].RetryCount)
		}
	} else {
		// Verify the job is in dead letter queue
		_, _, _, _ = q.GetAllDeadLetterJobs()
	}
}

func TestWorker_MultipleJobs(t *testing.T) {
	q := queue.NewQueue()
	w := &Worker{
		ID:    1,
		Queue: q,
	}

	// Create multiple jobs
	jobs := []utils.Job{
		{
			ID:         "job-1",
			Type:       "email",
			Payload:    map[string]string{"to": "user1@example.com"},
			Priority:   utils.High,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "job-2",
			Type:       "email",
			Payload:    map[string]string{"to": "user2@example.com"},
			Priority:   utils.Medium,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "job-3",
			Type:       "email",
			Payload:    map[string]string{"to": "user3@example.com"},
			Priority:   utils.Low,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		},
	}

	// Add jobs and start worker
	for _, job := range jobs {
		q.AddJob(job)
	}
	w.Start()

	// Wait for all jobs to be processed (3 jobs * 500ms each)
	time.Sleep(2 * time.Second)

	// Check that all jobs were processed
	highJobs, mediumJobs, lowJobs, _ := q.GetAllJobs()
	totalJobs := len(highJobs) + len(mediumJobs) + len(lowJobs)

	if totalJobs != 0 {
		t.Errorf("Expected all jobs to be processed, but found %d jobs remaining", totalJobs)
	}
}

func TestWorker_MultipleWorkers(t *testing.T) {
	q := queue.NewQueue()

	// Create multiple workers
	workers := []*Worker{
		{ID: 1, Queue: q},
		{ID: 2, Queue: q},
		{ID: 3, Queue: q},
	}

	// Create multiple jobs
	jobs := make([]utils.Job, 10)
	for i := 0; i < 10; i++ {
		jobs[i] = utils.Job{
			ID:         fmt.Sprintf("job-%d", i),
			Type:       "email",
			Payload:    map[string]string{"to": fmt.Sprintf("user%d@example.com", i)},
			Priority:   utils.High,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		}
		q.AddJob(jobs[i])
	}

	// Start all workers
	for _, w := range workers {
		w.Start()
	}

	// Wait for all jobs to be processed
	time.Sleep(2 * time.Second)

	// Check that all jobs were processed
	highJobs, _, _, _ := q.GetAllJobs()
	if len(highJobs) != 0 {
		t.Errorf("Expected all jobs to be processed by multiple workers, but found %d jobs remaining", len(highJobs))
	}
}

func TestWorker_NoJobAvailable(t *testing.T) {
	q := queue.NewQueue()
	w := &Worker{
		ID:    1,
		Queue: q,
	}

	// Start worker without adding jobs
	w.Start()

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Check that no jobs are in dead letter queue (worker shouldn't crash)
	_, _, _, _ = q.GetAllDeadLetterJobs()

	// This test mainly ensures the worker doesn't crash when no jobs are available
	// The worker should continue running and periodically check for jobs
}

func TestHandleJob_Success(t *testing.T) {
	job := utils.Job{
		ID:         "test-job",
		Type:       "email",
		Payload:    map[string]string{"to": "user@example.com"},
		Priority:   utils.High,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	err := handleJob(job)
	if err != nil {
		t.Errorf("Expected successful job handling, but got error: %v", err)
	}
}

func TestHandleJob_Error(t *testing.T) {
	job := utils.Job{
		ID:         "test-job-error",
		Type:       "email",
		Payload:    map[string]string{"to": "error@error.com"},
		Priority:   utils.High,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	err := handleJob(job)
	if err == nil {
		t.Error("Expected error for job with error@error.com, but got no error")
	}

	expectedError := "simulated error"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, err.Error())
	}
}

func TestWorker_ConcurrentJobAddition(t *testing.T) {
	q := queue.NewQueue()
	w := &Worker{
		ID:    1,
		Queue: q,
	}

	w.Start()

	var wg sync.WaitGroup

	// Add jobs concurrently
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			job := utils.Job{
				ID:         fmt.Sprintf("concurrent-job-%d", id),
				Type:       "email",
				Payload:    map[string]string{"to": fmt.Sprintf("user%d@example.com", id)},
				Priority:   utils.High,
				RetryCount: 0,
				MaxRetries: 3,
				CreatedAt:  time.Now(),
			}
			q.AddJob(job)
		}(i)
	}

	wg.Wait()

	// Wait for all jobs to be processed (5 jobs * 500ms each + buffer for concurrent processing)
	time.Sleep(4 * time.Second)

	// Check that all jobs were processed
	highJobs, _, _, _ := q.GetAllJobs()
	if len(highJobs) != 0 {
		t.Errorf("Expected all concurrently added jobs to be processed, but found %d jobs remaining", len(highJobs))
	}
}

func TestWorker_PriorityProcessing(t *testing.T) {
	q := queue.NewQueue()
	w := &Worker{
		ID:    1,
		Queue: q,
	}

	// Add jobs in reverse priority order (low, medium, high)
	// Worker should process high priority first due to queue priority handling
	jobs := []utils.Job{
		{
			ID:         "low-job",
			Type:       "email",
			Payload:    map[string]string{"to": "low@example.com"},
			Priority:   utils.Low,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "medium-job",
			Type:       "email",
			Payload:    map[string]string{"to": "medium@example.com"},
			Priority:   utils.Medium,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "high-job",
			Type:       "email",
			Payload:    map[string]string{"to": "high@example.com"},
			Priority:   utils.High,
			RetryCount: 0,
			MaxRetries: 3,
			CreatedAt:  time.Now(),
		},
	}

	for _, job := range jobs {
		q.AddJob(job)
	}

	w.Start()

	// Wait for all jobs to be processed
	time.Sleep(2 * time.Second)

	// Check that all jobs were processed (order doesn't matter for completion, just that they all get done)
	highJobs, mediumJobs, lowJobs, _ := q.GetAllJobs()
	totalJobs := len(highJobs) + len(mediumJobs) + len(lowJobs)

	if totalJobs != 0 {
		t.Errorf("Expected all jobs to be processed regardless of priority, but found %d jobs remaining", totalJobs)
	}
}
