package system

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMonitor(t *testing.T) {
	Convey("Given a new Monitor", t, func() {
		monitor := NewMonitor()

		Convey("When recording a metric", func() {
			labels := map[string]string{"process": "test"}
			monitor.RecordMetric(MetricTypeProcessDuration, 1.5, labels)

			Convey("Then it should be retrievable in the report", func() {
				metrics := monitor.Report()
				So(len(metrics), ShouldEqual, 1)
				So(metrics[0].Type, ShouldEqual, MetricTypeProcessDuration)
				So(metrics[0].Value, ShouldEqual, 1.5)
				So(metrics[0].Labels["process"], ShouldEqual, "test")
				So(metrics[0].Timestamp, ShouldHappenWithin, time.Second, time.Now())
			})
		})
	})
}
