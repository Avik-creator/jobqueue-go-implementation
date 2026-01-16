package queue

import (
	"errors"
	"sync"

	"github.com/Avik-creator/utils"
)

type JobQueue struct {
	mu              sync.Mutex
	queue           map[utils.Priority][]utils.Job
	deadLetterQueue map[utils.Priority][]utils.Job
}

func NewQueue() *JobQueue {
	return &JobQueue{
		queue: map[utils.Priority][]utils.Job{
			utils.High:   make([]utils.Job, 0),
			utils.Medium: make([]utils.Job, 0),
			utils.Low:    make([]utils.Job, 0),
		},
		deadLetterQueue: map[utils.Priority][]utils.Job{
			utils.High:   make([]utils.Job, 0),
			utils.Medium: make([]utils.Job, 0),
			utils.Low:    make([]utils.Job, 0),
		},
	}
}

func (q *JobQueue) AddJob(job utils.Job) *JobQueue {
	q.mu.Lock()
	defer q.mu.Unlock()

	switch job.Priority {
	case utils.High:
		q.queue[utils.High] = append(q.queue[utils.High], job)
	case utils.Medium:
		q.queue[utils.Medium] = append(q.queue[utils.Medium], job)
	case utils.Low:
		q.queue[utils.Low] = append(q.queue[utils.Low], job)
	}
	return q
}

func (q *JobQueue) GetJob() (utils.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.queue[utils.High]) > 0 {
		job := q.queue[utils.High][0]
		q.queue[utils.High] = q.queue[utils.High][1:]
		return job, nil
	}

	if len(q.queue[utils.Medium]) > 0 {
		job := q.queue[utils.Medium][0]
		q.queue[utils.Medium] = q.queue[utils.Medium][1:]
		return job, nil
	}

	if len(q.queue[utils.Low]) > 0 {
		job := q.queue[utils.Low][0]
		q.queue[utils.Low] = q.queue[utils.Low][1:]
		return job, nil
	}

	return utils.Job{}, errors.New("no job found")
}

func (q *JobQueue) RemoveJobFromQueue(job utils.Job) *JobQueue {
	q.mu.Lock()
	defer q.mu.Unlock()

	switch job.Priority {
	case utils.High:
		q.queue[utils.High] = utils.RemoveJob(q.queue[utils.High], job)
	case utils.Medium:
		q.queue[utils.Medium] = utils.RemoveJob(q.queue[utils.Medium], job)
	case utils.Low:
		q.queue[utils.Low] = utils.RemoveJob(q.queue[utils.Low], job)
	}
	return q
}

func (q *JobQueue) MoveJobToDeadLetterQueue(job utils.Job) *JobQueue {
	q.mu.Lock()
	defer q.mu.Unlock()

	// First remove from regular queue
	switch job.Priority {
	case utils.High:
		q.queue[utils.High] = utils.RemoveJob(q.queue[utils.High], job)
		q.deadLetterQueue[utils.High] = append(q.deadLetterQueue[utils.High], job)
	case utils.Medium:
		q.queue[utils.Medium] = utils.RemoveJob(q.queue[utils.Medium], job)
		q.deadLetterQueue[utils.Medium] = append(q.deadLetterQueue[utils.Medium], job)
	case utils.Low:
		q.queue[utils.Low] = utils.RemoveJob(q.queue[utils.Low], job)
		q.deadLetterQueue[utils.Low] = append(q.deadLetterQueue[utils.Low], job)
	}
	return q
}

func (q *JobQueue) GetDeadLetterJob() (utils.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.deadLetterQueue[utils.High]) > 0 {
		job := q.deadLetterQueue[utils.High][0]
		q.deadLetterQueue[utils.High] = q.deadLetterQueue[utils.High][1:]
		return job, nil
	}

	if len(q.deadLetterQueue[utils.Medium]) > 0 {
		job := q.deadLetterQueue[utils.Medium][0]
		q.deadLetterQueue[utils.Medium] = q.deadLetterQueue[utils.Medium][1:]
		return job, nil
	}

	if len(q.deadLetterQueue[utils.Low]) > 0 {
		job := q.deadLetterQueue[utils.Low][0]
		q.deadLetterQueue[utils.Low] = q.deadLetterQueue[utils.Low][1:]
		return job, nil
	}

	return utils.Job{}, errors.New("no job found")
}

func (q *JobQueue) GetAllJobs() ([]utils.Job, []utils.Job, []utils.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	highJobs := q.queue[utils.High]
	mediumJobs := q.queue[utils.Medium]
	lowJobs := q.queue[utils.Low]

	return highJobs, mediumJobs, lowJobs, nil
}

func (q *JobQueue) GetAllDeadLetterJobs() ([]utils.Job, []utils.Job, []utils.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	highJobs := q.deadLetterQueue[utils.High]
	mediumJobs := q.deadLetterQueue[utils.Medium]
	lowJobs := q.deadLetterQueue[utils.Low]

	return highJobs, mediumJobs, lowJobs, nil
}
