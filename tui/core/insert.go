// File: core/insert.go

package core

import "github.com/theapemachine/amsh/errnie"

// Insert mode implementation
type Insert struct {
	context *Context
}

func (i *Insert) Enter(ctx *Context) {
	errnie.Trace()
	i.context = ctx
	// Additional initialization if needed
}

func (i *Insert) Exit() {
	errnie.Trace()
	// Cleanup if needed
}
