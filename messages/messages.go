package messages

import (
	"github.com/theapemachine/amsh/components"
)

/*
MessageType allows components to evaluate whether or not they should process a message.
*/
type MessageType uint

const (
	// ComponentLoaded is sent when a component has been loaded.
	ComponentLoaded MessageType = iota
	// MessageOpenFile is sent when a file has been selected for editing.
	// It only sends the file path, the contents of the file still need to be loaded and read.
	MessageOpenFile
	// MessageWindowSize is sent when the window size changes.
	MessageWindowSize
	// MessageRender is send when a component wants to be rendered by the buffer.
	MessageRender
	// MessageMode is sent when the mode changes.
	MessageMode
	// MessageAnimate is a ticker message to animate objects.
	MessageAnimate
	// MessageError represents an error that happened.
	MessageError
)

/*
MessageContext allows the sender of a message to specify who should receive the message.
*/
type MessageContext uint

const (
	// All should be read by all components, active or inactive.
	All MessageContext = iota
	// Active should be read by all active components.
	Active
	// Inactive should be read by all inactive components.
	Inactive
	// Focused should be read by the currently focused component.
	Focused
)

/*
Message is a generic type that contains a message type, a context, and a data payload.
*/
type Message[T any] struct {
	Type    MessageType
	Context MessageContext
	Data    T
}

/*
NewMessage creates a new message with the given type, data, and context.
*/
func NewMessage[T any](msgType MessageType, data T, ctx MessageContext) Message[T] {
	return Message[T]{
		Type:    msgType,
		Context: ctx,
		Data:    data,
	}
}

func ShouldProcessMessage(componentState components.State, msgContext MessageContext) bool {
	switch msgContext {
	case All:
		return true
	case Active:
		return componentState == components.Active || componentState == components.Focused
	case Inactive:
		return componentState == components.Inactive
	case Focused:
		return componentState == components.Focused
	default:
		return false
	}
}
