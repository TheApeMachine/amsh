package provider

/*
Job is an interface that object can implement if they want to be
scheduled onto the worker pool.
*/
type Job interface {
	Execute(string) <-chan Event
}
