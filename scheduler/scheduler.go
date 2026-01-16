package scheduler

import (
	"container/heap"
	"sync"
	"time"

	"github.com/Avik-creator/queue"
	"github.com/Avik-creator/utils"
)

type ScheduleJob struct {
	Job          utils.Job
	ScheduleTime time.Time
	index        int
}

type JobHeap []*ScheduleJob

func (h JobHeap) Len() int           { return len(h) }
func (h JobHeap) Less(i, j int) bool { return h[i].ScheduleTime.Before(h[j].ScheduleTime) }

func (h JobHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i]; h[i].index, h[j].index = j, i }

func (h *JobHeap) Push(x interface{}) {
	job := x.(*ScheduleJob)
	job.index = len(*h)
	*h = append(*h, job)
}

func (h *JobHeap) Pop() interface{} {
	old := *h
	n := len(old)
	job := old[n-1]
	old[n-1] = nil // avoid memory leak
	*h = old[0 : n-1]
	return job
}

type Scheduler struct {
	mu    sync.Mutex
	heap  JobHeap
	queue *queue.JobQueue
}

func NewScheduler(q *queue.JobQueue) *Scheduler {
	h := make(JobHeap, 0)
	heap.Init(&h)

	s := &Scheduler{heap: h, queue: q}
	go s.poller()
	return s
}

func (s *Scheduler) Scheduler(j utils.Job, delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	scheduled := &ScheduleJob{
		Job:          j,
		ScheduleTime: time.Now().Add(delay),
	}

	heap.Push(&s.heap, scheduled)
}

func (s *Scheduler) poller() {
	for {
		s.mu.Lock()
		if s.heap.Len() > 0 {
			next := s.heap[0]
			if time.Now().After(next.ScheduleTime) {
				heap.Pop(&s.heap)
				s.queue.AddJob(next.Job)
			}
		}

		s.mu.Unlock()
		time.Sleep(500 * time.Millisecond)
	}
}
