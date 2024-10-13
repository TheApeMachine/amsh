package twoface

import (
	"io"

	"github.com/theapemachine/amsh/errnie"
)

/*
Worker is a concurrent context for executing jobs.
It implements the io.ReadCloser interface, to maintain a high level of compatibility with
other components, which mostly all implement the io.ReadWriteCloser interface.
*/
type Worker struct {
	ID         int
	JobChannel chan Job
	WorkerPool chan chan Job
	quit       chan struct{}
	pr         *io.PipeReader
	pw         *io.PipeWriter
}

/*
NewWorker instantiates a goroutine that will execute jobs.
*/
func NewWorker(id int, workerPool chan chan Job) *Worker {
	pr, pw := io.Pipe()

	return &Worker{
		ID:         id,
		JobChannel: make(chan Job),
		WorkerPool: workerPool,
		quit:       make(chan struct{}),
		pr:         pr,
		pw:         pw,
	}
}

/*
Read implements the io.Reader interface, and is used to execute jobs.
Think of it as 'lazy' execution, where the write action loads the job onto the worker,
and the read action is called by the pool to get the results of the job.
This means that if the job is cancelled before read is called, we have not wasted any resources.
*/
func (w *Worker) Read(p []byte) (n int, err error) {
	go func() {
		for {
			// Register the worker's job channel into the worker pool to signal that we
			// are ready to accept a new job.
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				// We have received a job, which we will execute by reading from it, given that
				// jobs are also only executed by reading from them.
				_, err = io.Copy(w.pw, job)
				errnie.Error(err)
			case <-w.quit:
				// We have been asked to stop, so we close the pipe to signal EOF to the reader.
				w.pw.CloseWithError(io.EOF)
				return
			}
		}
	}()

	return w.pr.Read(p)
}

/*
Close the worker, which will signal EOF to any readers.
*/
func (w *Worker) Close() error {
	close(w.quit)
	return nil
}
