package twoface

import (
	"io"
)

/*
Job is an interface any type can implement if they want to be able to use the worker pool.
*/
type Job interface {
	io.ReadWriteCloser
}

/*
NewJob is a convenience method to convert any incoming structured type to a Job interface.
*/
func NewJob(jobType Job) Job {
	return jobType
}
