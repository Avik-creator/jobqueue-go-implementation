# Job Queue System

A high-performance, priority-based job queue system built in Go with worker pools, scheduling capabilities, and a command-line interface.

## Features

- **Priority-based Queue**: Three priority levels (High, Medium, Low) with FIFO ordering within each priority
- **Worker Pool**: Concurrent job processing with configurable worker count
- **Retry Mechanism**: Exponential backoff retry logic for failed jobs
- **Dead Letter Queue**: Automatic handling of jobs that exceed maximum retry attempts
- **Job Scheduling**: Schedule jobs to run at a future time with delays
- **CLI Interface**: Easy-to-use command-line interface for queue management
- **Thread-safe Operations**: Mutex-protected concurrent access to queue operations

## Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   CLI       │    │  Scheduler  │    │   Workers   │
│  Interface  │───▶│             │───▶│             │
└─────────────┘    └─────────────┘    └─────────────┘
        │                   │                   │
        ▼                   ▼                   ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Queue     │    │   Queue     │    │   Queue     │
│ Management │    │ Management │    │ Management │
└─────────────┘    └─────────────┘    └─────────────┘
```

### Components

- **Queue**: Priority-based job storage and retrieval
- **Worker**: Job processing with retry logic and error handling
- **Scheduler**: Delayed job execution with heap-based timing
- **CLI**: Command-line interface for queue operations

## Installation

### Prerequisites

- Go 1.19 or later

### Build from Source

```bash
git clone <repository-url>
cd jobQueue
go mod tidy
go build -o jobqueue .
```

## Usage

### Command Line Interface

The CLI provides easy management of jobs, workers, and queues.

#### Enqueue a Job

```bash
# Enqueue a high-priority job
./jobqueue enqueue --to user@example.com --priority high

# Enqueue a job with custom retry count
./jobqueue enqueue --to user@example.com --retries 5

# Schedule a job to run after 60 seconds
./jobqueue enqueue --to user@example.com --delay 60
```

#### Start Workers

```bash
# Start a single worker
./jobqueue start

# Start multiple workers
./jobqueue start --count 5
```

#### View Dead Letter Queue

```bash
# Show failed jobs
./jobqueue dlq
```

### Programmatic Usage

```go
package main

import (
    "time"
    "github.com/Avik-creator/queue"
    "github.com/Avik-creator/worker"
    "github.com/Avik-creator/utils"
)

// Create a queue
q := queue.NewQueue()

// Create a job
job := utils.Job{
    ID:         "job-123",
    Type:       "email",
    Payload:    map[string]string{"to": "user@example.com"},
    Priority:   utils.High,
    MaxRetries: 3,
    CreatedAt:  time.Now(),
}

// Add job to queue
q.AddJob(job)

// Start a worker
w := &worker.Worker{ID: 1, Queue: q}
w.Start()
```

## CLI Commands

### `enqueue`

Enqueue a new job into the queue.

```bash
./jobqueue enqueue [flags]
```

**Flags:**
- `--to string`: Recipient email address (required)
- `--priority string`: Job priority (low, medium, high) (default "low")
- `--retries int`: Maximum retry attempts (default 3)
- `--delay int`: Delay in seconds before execution (default 0)

### `start`

Start one or more workers to process jobs.

```bash
./jobqueue start [flags]
```

**Flags:**
- `--count int`: Number of workers to start (default 1)

### `dlq`

Display jobs in the dead letter queue.

```bash
./jobqueue dlq
```

## Job Processing

### Job States

1. **Queued**: Job is waiting in the priority queue
2. **Processing**: Job is being processed by a worker
3. **Completed**: Job finished successfully
4. **Failed**: Job failed and will be retried
5. **Dead Letter**: Job exceeded maximum retries and moved to dead letter queue

### Retry Logic

Jobs that fail during processing are automatically retried with exponential backoff:

- 1st retry: 2 seconds delay
- 2nd retry: 4 seconds delay
- 3rd retry: 8 seconds delay
- etc.

After reaching `MaxRetries`, the job is moved to the dead letter queue.

## Configuration

### Job Priorities

- **High**: Processed first
- **Medium**: Processed after high priority jobs
- **Low**: Processed last

### Default Settings

- Default priority: Low
- Default max retries: 3
- Worker sleep interval: 1 second
- Scheduler poll interval: 500ms

## Testing

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./queue -v
go test ./worker -v
```

## Examples

### Basic Email Job Processing

```bash
# Terminal 1: Start 2 workers
./jobqueue start --count 2

# Terminal 2: Enqueue jobs
./jobqueue enqueue --to user1@example.com --priority high
./jobqueue enqueue --to user2@example.com --priority medium
./jobqueue enqueue --to user3@example.com --priority low

# Terminal 3: Check dead letter queue
./jobqueue dlq
```

### Scheduled Jobs

```bash
# Schedule a job to run in 30 seconds
./jobqueue enqueue --to future@example.com --delay 30 --priority high

# Start workers to process when ready
./jobqueue start
```

## API Reference

### Queue Package

- `NewQueue() *JobQueue`: Create a new job queue
- `AddJob(job Job)`: Add a job to the queue
- `GetJob() (Job, error)`: Retrieve next job by priority
- `GetAllJobs() ([]Job, []Job, []Job, error)`: Get all jobs by priority
- `GetAllDeadLetterJobs() ([]Job, []Job, []Job, error)`: Get dead letter jobs

### Worker Package

- `Start()`: Begin processing jobs from the queue
- `handleJob(job Job) error`: Process individual job (internal)

### Scheduler Package

- `NewScheduler(queue *JobQueue) *Scheduler`: Create a new scheduler
- `Scheduler(job Job, delay time.Duration)`: Schedule a job for future execution

### Utils Package

- `Job`: Job structure with ID, type, payload, priority, retry info
- `Priority`: Priority enumeration (High, Medium, Low)
- `RemoveJob(jobs []Job, job Job) []Job`: Remove job from slice

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Performance Notes

- The queue uses mutexes for thread-safe operations
- Workers continuously poll the queue for new jobs
- Scheduler uses a heap for efficient delayed job management
- Exponential backoff prevents system overload during failures