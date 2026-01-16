package queue

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/Avik-creator/utils"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue()

	if q == nil {
		t.Fatal("NewQueue returned nil")
	}

	if q.queue == nil {
		t.Fatal("queue map not initialized")
	}

	if q.deadLetterQueue == nil {
		t.Fatal("deadLetterQueue map not initialized")
	}

	// Check that all priority queues are initialized
	expectedPriorities := []utils.Priority{utils.High, utils.Medium, utils.Low}
	for _, priority := range expectedPriorities {
		if _, exists := q.queue[priority]; !exists {
			t.Errorf("queue not initialized for priority %v", priority)
		}
		if _, exists := q.deadLetterQueue[priority]; !exists {
			t.Errorf("deadLetterQueue not initialized for priority %v", priority)
		}
		if len(q.queue[priority]) != 0 {
			t.Errorf("queue for priority %v should be empty, got %d items", priority, len(q.queue[priority]))
		}
		if len(q.deadLetterQueue[priority]) != 0 {
			t.Errorf("deadLetterQueue for priority %v should be empty, got %d items", priority, len(q.deadLetterQueue[priority]))
		}
	}
}

func TestAddJob(t *testing.T) {
	q := NewQueue()

	// Create test jobs
	job1 := utils.Job{
		ID:         "job1",
		Type:       "test",
		Payload:    map[string]string{"key": "value"},
		Priority:   utils.High,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	job2 := utils.Job{
		ID:         "job2",
		Type:       "test",
		Payload:    map[string]string{"key": "value2"},
		Priority:   utils.Medium,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	job3 := utils.Job{
		ID:         "job3",
		Type:       "test",
		Payload:    map[string]string{"key": "value3"},
		Priority:   utils.Low,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	// Add jobs
	q.AddJob(job1)
	q.AddJob(job2)
	q.AddJob(job3)

	// Verify jobs were added to correct queues
	if len(q.queue[utils.High]) != 1 {
		t.Errorf("Expected 1 high priority job, got %d", len(q.queue[utils.High]))
	}
	if len(q.queue[utils.Medium]) != 1 {
		t.Errorf("Expected 1 medium priority job, got %d", len(q.queue[utils.Medium]))
	}
	if len(q.queue[utils.Low]) != 1 {
		t.Errorf("Expected 1 low priority job, got %d", len(q.queue[utils.Low]))
	}

	// Verify job content
	if q.queue[utils.High][0].ID != "job1" {
		t.Errorf("Expected job1 in high priority queue, got %s", q.queue[utils.High][0].ID)
	}
	if q.queue[utils.Medium][0].ID != "job2" {
		t.Errorf("Expected job2 in medium priority queue, got %s", q.queue[utils.Medium][0].ID)
	}
	if q.queue[utils.Low][0].ID != "job3" {
		t.Errorf("Expected job3 in low priority queue, got %s", q.queue[utils.Low][0].ID)
	}
}

func TestGetJob(t *testing.T) {
	q := NewQueue()

	// Test empty queue
	_, err := q.GetJob()
	if err == nil {
		t.Error("Expected error when getting job from empty queue")
	}
	if err.Error() != "no job found" {
		t.Errorf("Expected 'no job found' error, got '%s'", err.Error())
	}

	// Add jobs with different priorities
	job1 := utils.Job{ID: "job1", Priority: utils.Low, CreatedAt: time.Now()}
	job2 := utils.Job{ID: "job2", Priority: utils.Medium, CreatedAt: time.Now()}
	job3 := utils.Job{ID: "job3", Priority: utils.High, CreatedAt: time.Now()}

	q.AddJob(job1)
	q.AddJob(job2)
	q.AddJob(job3)

	// Should get high priority job first
	job, err := q.GetJob()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if job.ID != "job3" {
		t.Errorf("Expected job3 (high priority), got %s", job.ID)
	}

	// Should get medium priority job next
	job, err = q.GetJob()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if job.ID != "job2" {
		t.Errorf("Expected job2 (medium priority), got %s", job.ID)
	}

	// Should get low priority job last
	job, err = q.GetJob()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if job.ID != "job1" {
		t.Errorf("Expected job1 (low priority), got %s", job.ID)
	}

	// Queue should be empty now
	_, err = q.GetJob()
	if err == nil {
		t.Error("Expected error when queue is empty")
	}
}

func TestRemoveJobFromQueue(t *testing.T) {
	q := NewQueue()

	job1 := utils.Job{ID: "job1", Priority: utils.High, CreatedAt: time.Now()}
	job2 := utils.Job{ID: "job2", Priority: utils.High, CreatedAt: time.Now()}
	job3 := utils.Job{ID: "job3", Priority: utils.Medium, CreatedAt: time.Now()}

	q.AddJob(job1)
	q.AddJob(job2)
	q.AddJob(job3)

	// Verify jobs added
	if len(q.queue[utils.High]) != 2 {
		t.Errorf("Expected 2 high priority jobs, got %d", len(q.queue[utils.High]))
	}
	if len(q.queue[utils.Medium]) != 1 {
		t.Errorf("Expected 1 medium priority job, got %d", len(q.queue[utils.Medium]))
	}

	// Remove job1 from high priority queue
	q.RemoveJobFromQueue(job1)

	if len(q.queue[utils.High]) != 1 {
		t.Errorf("Expected 1 high priority job after removal, got %d", len(q.queue[utils.High]))
	}
	if q.queue[utils.High][0].ID != "job2" {
		t.Errorf("Expected job2 to remain in high priority queue, got %s", q.queue[utils.High][0].ID)
	}

	// Medium priority queue should remain unchanged
	if len(q.queue[utils.Medium]) != 1 {
		t.Errorf("Expected 1 medium priority job to remain unchanged, got %d", len(q.queue[utils.Medium]))
	}
}

func TestMoveJobToDeadLetterQueue(t *testing.T) {
	q := NewQueue()

	job1 := utils.Job{ID: "job1", Priority: utils.High, CreatedAt: time.Now()}
	job2 := utils.Job{ID: "job2", Priority: utils.Medium, CreatedAt: time.Now()}

	q.AddJob(job1)
	q.AddJob(job2)

	// Move job1 to dead letter queue
	q.MoveJobToDeadLetterQueue(job1)

	// Verify job1 is in dead letter queue
	if len(q.deadLetterQueue[utils.High]) != 1 {
		t.Errorf("Expected 1 job in high priority dead letter queue, got %d", len(q.deadLetterQueue[utils.High]))
	}
	if q.deadLetterQueue[utils.High][0].ID != "job1" {
		t.Errorf("Expected job1 in dead letter queue, got %s", q.deadLetterQueue[utils.High][0].ID)
	}

	// Verify job1 is removed from regular queue
	if len(q.queue[utils.High]) != 0 {
		t.Errorf("Expected 0 jobs in high priority queue after move, got %d", len(q.queue[utils.High]))
	}

	// Verify job2 remains in regular queue
	if len(q.queue[utils.Medium]) != 1 {
		t.Errorf("Expected 1 job in medium priority queue, got %d", len(q.queue[utils.Medium]))
	}
}

func TestGetDeadLetterJob(t *testing.T) {
	q := NewQueue()

	// Test empty dead letter queue
	_, err := q.GetDeadLetterJob()
	if err == nil {
		t.Error("Expected error when getting job from empty dead letter queue")
	}

	// Add jobs to dead letter queue
	job1 := utils.Job{ID: "job1", Priority: utils.Low, CreatedAt: time.Now()}
	job2 := utils.Job{ID: "job2", Priority: utils.Medium, CreatedAt: time.Now()}
	job3 := utils.Job{ID: "job3", Priority: utils.High, CreatedAt: time.Now()}

	q.MoveJobToDeadLetterQueue(job1)
	q.MoveJobToDeadLetterQueue(job2)
	q.MoveJobToDeadLetterQueue(job3)

	// Should get high priority job first
	job, err := q.GetDeadLetterJob()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if job.ID != "job3" {
		t.Errorf("Expected job3 (high priority), got %s", job.ID)
	}

	// Should get medium priority job next
	job, err = q.GetDeadLetterJob()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if job.ID != "job2" {
		t.Errorf("Expected job2 (medium priority), got %s", job.ID)
	}

	// Should get low priority job last
	job, err = q.GetDeadLetterJob()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if job.ID != "job1" {
		t.Errorf("Expected job1 (low priority), got %s", job.ID)
	}
}

func TestGetAllJobs(t *testing.T) {
	q := NewQueue()

	job1 := utils.Job{ID: "job1", Priority: utils.High, CreatedAt: time.Now()}
	job2 := utils.Job{ID: "job2", Priority: utils.Medium, CreatedAt: time.Now()}
	job3 := utils.Job{ID: "job3", Priority: utils.Low, CreatedAt: time.Now()}
	job4 := utils.Job{ID: "job4", Priority: utils.High, CreatedAt: time.Now()}

	q.AddJob(job1)
	q.AddJob(job2)
	q.AddJob(job3)
	q.AddJob(job4)

	highJobs, mediumJobs, lowJobs, err := q.GetAllJobs()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(highJobs) != 2 {
		t.Errorf("Expected 2 high priority jobs, got %d", len(highJobs))
	}
	if len(mediumJobs) != 1 {
		t.Errorf("Expected 1 medium priority job, got %d", len(mediumJobs))
	}
	if len(lowJobs) != 1 {
		t.Errorf("Expected 1 low priority job, got %d", len(lowJobs))
	}

	// Verify job IDs
	expectedHigh := []string{"job1", "job4"}
	actualHigh := []string{highJobs[0].ID, highJobs[1].ID}
	if !reflect.DeepEqual(expectedHigh, actualHigh) {
		t.Errorf("Expected high priority jobs %v, got %v", expectedHigh, actualHigh)
	}

	if mediumJobs[0].ID != "job2" {
		t.Errorf("Expected job2 in medium priority jobs, got %s", mediumJobs[0].ID)
	}

	if lowJobs[0].ID != "job3" {
		t.Errorf("Expected job3 in low priority jobs, got %s", lowJobs[0].ID)
	}
}

func TestGetAllDeadLetterJobs(t *testing.T) {
	q := NewQueue()

	job1 := utils.Job{ID: "job1", Priority: utils.High, CreatedAt: time.Now()}
	job2 := utils.Job{ID: "job2", Priority: utils.Medium, CreatedAt: time.Now()}

	q.MoveJobToDeadLetterQueue(job1)
	q.MoveJobToDeadLetterQueue(job2)

	highJobs, mediumJobs, lowJobs, err := q.GetAllDeadLetterJobs()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(highJobs) != 1 {
		t.Errorf("Expected 1 high priority dead letter job, got %d", len(highJobs))
	}
	if len(mediumJobs) != 1 {
		t.Errorf("Expected 1 medium priority dead letter job, got %d", len(mediumJobs))
	}
	if len(lowJobs) != 0 {
		t.Errorf("Expected 0 low priority dead letter jobs, got %d", len(lowJobs))
	}

	if highJobs[0].ID != "job1" {
		t.Errorf("Expected job1 in high priority dead letter jobs, got %s", highJobs[0].ID)
	}
	if mediumJobs[0].ID != "job2" {
		t.Errorf("Expected job2 in medium priority dead letter jobs, got %s", mediumJobs[0].ID)
	}
}

func TestConcurrencySafety(t *testing.T) {
	q := NewQueue()
	const numGoroutines = 10
	const jobsPerGoroutine = 100

	var wg sync.WaitGroup

	// Add jobs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < jobsPerGoroutine; j++ {
				job := utils.Job{
					ID:        fmt.Sprintf("job-%d-%d", id, j),
					Priority:  utils.High,
					CreatedAt: time.Now(),
				}
				q.AddJob(job)
			}
		}(i)
	}

	wg.Wait()

	// Verify all jobs were added
	totalJobs := 0
	for _, jobs := range q.queue {
		totalJobs += len(jobs)
	}

	expectedTotal := numGoroutines * jobsPerGoroutine
	if totalJobs != expectedTotal {
		t.Errorf("Expected %d total jobs, got %d", expectedTotal, totalJobs)
	}

	// Remove jobs concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < jobsPerGoroutine; j++ {
				job := utils.Job{
					ID:       fmt.Sprintf("job-%d-%d", id, j),
					Priority: utils.High,
				}
				q.RemoveJobFromQueue(job)
			}
		}(i)
	}

	wg.Wait()

	// Verify all jobs were removed
	totalJobs = 0
	for _, jobs := range q.queue {
		totalJobs += len(jobs)
	}

	if totalJobs != 0 {
		t.Errorf("Expected 0 total jobs after removal, got %d", totalJobs)
	}
}

func TestMethodChaining(t *testing.T) {
	q := NewQueue()

	job1 := utils.Job{ID: "job1", Priority: utils.High, CreatedAt: time.Now()}
	job2 := utils.Job{ID: "job2", Priority: utils.Medium, CreatedAt: time.Now()}

	// Test method chaining
	result := q.AddJob(job1).AddJob(job2).RemoveJobFromQueue(job1).MoveJobToDeadLetterQueue(job2)

	if result != q {
		t.Error("Method chaining should return the same queue instance")
	}

	// Verify final state
	if len(q.queue[utils.High]) != 0 {
		t.Error("High priority queue should be empty after removing job1")
	}
	if len(q.deadLetterQueue[utils.Medium]) != 1 {
		t.Error("Medium priority dead letter queue should have 1 job")
	}
}
