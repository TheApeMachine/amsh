package twoface

import "context"

type Job interface {
	Process(ctx context.Context) error
}
