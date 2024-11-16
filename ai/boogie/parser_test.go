package boogie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParser(t *testing.T) {
	Convey("Given a boogie program", t, func() {
		program := `
		out <= () <= in
		`

		_ = program
	})
}
