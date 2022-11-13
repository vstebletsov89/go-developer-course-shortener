package worker

import (
	"context"
	"go-developer-course-shortener/internal/app/repository"
	"testing"
)

func TestNewWorkerPool(t *testing.T) {
	tests := []struct {
		name string
		repo repository.Repository
		jobs <-chan Job
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
		name         string
		repo         repository.Repository
		cancel       bool
		awaitingJobs bool
	}{
		{
			name:         "workerPool.Run emulation of shutdown without awaiting jobs",
			repo:         repository.NewInMemoryRepository(),
			cancel:       true,
			awaitingJobs: false,
		},
		{
			name:         "workerPool.Run emulation of shutdown with awaiting jobs",
			repo:         repository.NewInMemoryRepository(),
			cancel:       true,
			awaitingJobs: true,
		},
		{
			name:         "workerPool.Run without shutdown",
			repo:         repository.NewInMemoryRepository(),
			cancel:       false,
			awaitingJobs: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs := make(chan Job, MaxWorkerPoolSize)
			workerPool := NewWorkerPool(tt.repo, jobs)

			var ctx context.Context
			var cancel context.CancelFunc

			if tt.cancel {
				ctx, cancel = context.WithCancel(context.Background())
				cancel() // cancel context
			} else {
				ctx = context.Background()
			}

			go workerPool.Run(ctx)

			if tt.awaitingJobs {
				for i := 0; i < 10000; i++ {
					j := Job{UserID: "testUser", ShortURLS: nil}
					jobs <- j
				}
			}

			ctx.Done()
		})
	}
}
