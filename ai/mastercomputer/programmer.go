package mastercomputer

import "context"

type Programmer struct {
	agent *Agent
}

func NewProgrammer(ctx context.Context) *Programmer {
	return &Programmer{
		agent: NewAgent(ctx, "programmer"),
	}
}
