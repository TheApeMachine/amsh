package system

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProcessManager(t *testing.T) {
	Convey("Given a new ProcessManager", t, func() {
		pm := NewProcessManager(&Architecture{})
		_ = pm
	})
}
