// File: core/normal.go

package core

import "github.com/theapemachine/amsh/errnie"

// Normal mode implementation
type Normal struct {
	context *Context
}

func (n *Normal) Enter(ctx *Context) {
	errnie.Trace()
	n.context = ctx
	// Additional initialization if needed
}

func (n *Normal) Exit() {
	errnie.Trace()
	// Cleanup if needed
}
