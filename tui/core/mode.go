package core

type Mode interface {
	Enter(context *Context)
	Exit()
}
