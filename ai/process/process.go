package process

import "github.com/theapemachine/amsh/ai/process/fractal"

type Process interface {
	GenerateSchema() string
}

var ProcessMap = map[string]Process{
	"fractal_pattern": &fractal.Process{},
}
