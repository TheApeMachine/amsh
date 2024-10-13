package twoface

import (
	"context"
	"sync"
	"time"
)

type Scaler struct {
	pool       *Pool
	maxWorkers int
	minWorkers int
	interval   time.Duration
	mu         sync.Mutex
}

func NewScaler(pool *Pool, minWorkers, maxWorkers int, interval time.Duration) *Scaler {
	return &Scaler{
		pool:       pool,
		minWorkers: minWorkers,
		maxWorkers: maxWorkers,
		interval:   interval,
	}
}

func (s *Scaler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-ticker.C:
			s.adjustWorkers()
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (s *Scaler) adjustWorkers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	queueLength := len(s.pool.JobQueue)
	currentWorkers := len(s.pool.workers)

	if queueLength > currentWorkers && currentWorkers < s.maxWorkers {
		// Increase workers.
		s.pool.addWorker()
	} else if queueLength < currentWorkers && currentWorkers > s.minWorkers {
		// Decrease workers.
		s.pool.removeWorker()
	}
}
