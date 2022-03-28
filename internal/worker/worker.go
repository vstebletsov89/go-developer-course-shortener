package worker

import (
	"context"
	"go-developer-course-shortener/internal/app/repository"
	"log"
	"sync"
)

const MaxWorkerPoolSize = 10

type Job struct {
	UserID    string
	ShortURLS []string
}

type Pool struct {
	repository repository.Repository
	inputCh    <-chan Job
}

func NewWorkerPool(repo repository.Repository, inputCh <-chan Job) *Pool {
	return &Pool{repository: repo, inputCh: inputCh}
}

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
			return
		}
	}
}
