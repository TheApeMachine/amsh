package twoface

import "io"

type Job interface {
	io.ReadWriteCloser
}
