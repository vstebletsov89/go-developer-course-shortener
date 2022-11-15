package worker

import (
	"context"
	"fmt"
	"go-developer-course-shortener/internal/app/repository"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name string
		repo repository.Repository
		jobs chan Job
	}{
		{
			name: "test NewWorkerPool",
			repo: repository.NewInMemoryRepository(),
			jobs: make(chan Job, MaxWorkerPoolSize),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWorkerPool(tt.repo, tt.jobs)
			if got == nil {
				t.Errorf("NewWorkerPool() is nil")
				return
			}
		})
	}
}

func TestPoolRun(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		shutdown bool
	}{
		{
			name:     "workerPool.Run normal execution",
			userID:   "UserNormal",
			shutdown: false,
		},
		{
			name:     "workerPool.Run emulation of shutdown",
			userID:   "UserShutdown",
			shutdown: true,
		},
	}

	repo := repository.NewInMemoryRepository()

	jobs := make(chan Job, MaxWorkerPoolSize)
	workerPool := NewWorkerPool(repo, jobs)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go workerPool.Run(ctx)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < 3; i++ {
				if tt.shutdown {
					cancel()
					// send one task and close the channel
					user := fmt.Sprintf("%v%d", tt.userID, i)
					j := Job{UserID: user, ShortURLS: nil}
					jobs <- j
					workerPool.ClosePool()
					break
				} else {
					user := fmt.Sprintf("%v%d", tt.userID, i)
					j := Job{UserID: user, ShortURLS: nil}
					jobs <- j
				}
			}

			time.Sleep(1 * time.Second) // server simulation
		})
	}
}
