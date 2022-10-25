// Package worker provides primitives for concurrency.
package worker

import (
	"context"
	"go-developer-course-shortener/internal/app/repository"
	"log"
	"sync"
)

// MaxWorkerPoolSize maximum size of workers for parallel requests.
const MaxWorkerPoolSize = 10

// Job is a task to be executed.
type Job struct {
	// user id for current request.
	UserID string
	// slice of short urls to be deleted.
	ShortURLS []string
}

// Pool represents queue of jobs.
type Pool struct {
	repository repository.Repository
	inputCh    <-chan Job
}

// NewWorkerPool returns a new Pool, serving the provided Repository.
func NewWorkerPool(repo repository.Repository, inputCh <-chan Job) *Pool {
	return &Pool{repository: repo, inputCh: inputCh}
}

// Run processing Job channels in the current context.
func (p *Pool) Run(ctx context.Context) {
	wg := sync.WaitGroup{}
	for {
		select {
		case v := <-p.inputCh:
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := p.repository.DeleteURLS(ctx, v.UserID, v.ShortURLS); err != nil {
					log.Println(err)
					return
				}
			}()
		case <-ctx.Done():
			log.Println("Worker pool context done")
			return
		}
	}
}
