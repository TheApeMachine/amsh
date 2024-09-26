package twoface

import (
	"time"
)

// Scaler controls the size of a worker pool and dynamically scales the amount of worker routines.
type Scaler struct {
	interval time.Duration
	rate     int
	stats    int64
	period   int
	level    int
	samples  int
	overload bool
	lower    bool
	pool     *Pool
	maxIdle  time.Duration
}

// NewScaler constructs a scaler which controls the size of a worker pool dynamically.
func NewScaler(pool *Pool) *Scaler {
	return &Scaler{
		interval: 100,
		rate:     10,
		stats:    0,
		period:   0,
		level:    1,
		samples:  3,
		overload: false,
		lower:    false,
		pool:     pool,
		maxIdle:  1 * time.Second,
	}
}

// Run starts the scaler to periodically evaluate and adjust the worker pool size.
func (scaler *Scaler) Run() {
	ticker := time.NewTicker(scaler.interval * time.Millisecond)

	go func() {
		for {
			select {
			case <-scaler.pool.ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				scaler.load()
				if !scaler.overload && len(scaler.pool.jobQueue) > 0 {
					scaler.Grow()
				}
				if scaler.overload {
					scaler.Shrink()
				}
			}
		}
	}()
}

// load determines if the pool's performance is degrading.
func (scaler *Scaler) load() {
	scaler.period++
	prev := scaler.stats
	scaler.stats = 0

	var count int
	for _, worker := range scaler.pool.workers {
		if worker.lastDuration != 0 {
			scaler.stats += worker.lastDuration
			count++
		}
	}

	if prev == 0 || scaler.stats == 0 {
		return
	}

	scaler.stats = scaler.stats / int64(count)
	lower := scaler.stats > prev

	if scaler.lower != lower {
		scaler.period = 0
		if scaler.level > 1 {
			scaler.level--
		}
	}

	scaler.lower = lower

	if scaler.period >= scaler.samples {
		scaler.period = 0
		if scaler.level < 3 {
			scaler.level++
		}

		if scaler.stats > prev {
			scaler.overload = true
			return
		}

		scaler.overload = false
		return
	}
}

// Grow increases the size of the worker pool.
func (scaler *Scaler) Grow() {
	if !scaler.overload {
		for i := 0; i < scaler.rate*scaler.level; i++ {
			scaler.pool.workers = append(scaler.pool.workers, NewWorker(
				len(scaler.pool.workers), scaler.pool.workerPool, scaler.pool.ctx,
			))
		}
	}
}

// Shrink reduces the size of the worker pool.
func (scaler *Scaler) Shrink() {
	if len(scaler.pool.workers) == 0 {
		return
	}

	if scaler.overload {
		for i := 0; i < scaler.rate; i++ {
			scaler.drain(scaler.pool.workers[i], i)
		}
		return
	}

	for idx, worker := range scaler.pool.workers {
		if time.Since(worker.lastUse) > scaler.maxIdle {
			scaler.drain(worker, idx)
		}
	}
}

func (scaler *Scaler) drain(worker *Worker, i int) {
	worker.Close()
	scaler.pool.workers = append(scaler.pool.workers[:i], scaler.pool.workers[i+1:]...)
}
