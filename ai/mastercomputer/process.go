package mastercomputer

import "github.com/theapemachine/amsh/ai/boogie"

/*
Process defines an interface that object can implement if the want to act
as a predefined process. Predefined processes are used to direct specific
behavior, useful is cases where we know what should be done based on an input.
*/
type Process interface {
	NextState(
		boogie.Operation, string, boogie.State,
	) boogie.State
}
