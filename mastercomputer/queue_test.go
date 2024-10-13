package mastercomputer

import (
	"context"
	"io"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/theapemachine/amsh/data"
)

type MockWorker struct {
	Worker
}

func NewMockWorker(ctx context.Context, artifact data.Artifact) *MockWorker {
	return &MockWorker{*NewWorker(ctx, artifact)}
}

func TestWorkerIntegration(t *testing.T) {
	Convey("Given a worker and a queue", t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		publisher := NewMockWorker(ctx, data.New("test", "test", "test", []byte{}))
		publisher.ID = "test-publisher"

		Convey("When registering a worker", func() {
			worker := NewWorker(ctx, data.New("test", "test", "test", []byte{}))
			worker.Initialize()

			_, err := io.Copy(worker, data.New(
				"test", "initialize", "reasoner", nil,
			).Poke(
				"format", "reasoning",
			).Poke(
				"toolset", "reasoning",
			))

			Convey("It should initialize successfully", func() {
				So(err, ShouldBeNil)
				So(worker, ShouldNotBeNil)
				So(worker.OK, ShouldBeTrue)
				So(worker.State, ShouldEqual, WorkerStateReady)
			})
		})
	})
}
