package worker

import (
	"context"
	"go-developer-course-shortener/internal/app/repository"
	"testing"
	"time"
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
		name    string
		repo    repository.Repository
		timeout bool
	}{
		{
			name:    "test workerPool.Run with timeout",
			repo:    repository.NewInMemoryRepository(),
			timeout: true,
		},
		{
			name:    "test workerPool.Run without timeout",
			repo:    repository.NewInMemoryRepository(),
			timeout: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobs := make(chan Job, MaxWorkerPoolSize)
			workerPool := NewWorkerPool(tt.repo, jobs)

			if tt.timeout {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				defer cancel()
				go workerPool.Run(ctx)
			} else {
				ctx := context.Background()
				go workerPool.Run(ctx)

				j := Job{UserID: "testUser", ShortURLS: nil}
				jobs <- j
			}
		})
	}
}
