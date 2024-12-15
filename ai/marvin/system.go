package marvin

import (
	"context"

	"github.com/theapemachine/amsh/twoface"
)

type System struct {
	accumulator *twoface.Accumulator
}

func NewSystem() *System {
	return &System{
		accumulator: twoface.NewAccumulator(),
	}
}

func (system *System) Generate(user *data.Data) (err error) {
	agent := provider.NewAgent(context.Background(), "test")

	go func() {
		if _, err = io.Copy(agent, user); err != nil {
			errnie.Error(err)
		}
	}()
}
