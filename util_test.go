package instruments

import (
	"github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
)

var _ = ginkgo.Describe("MetricID", func() {
	ginkgo.DescribeTable("should assemble",
		func(name string, tags []string, x string) {
			m := MetricID(name, tags)
			Expect(string(m)).To(Equal(x))
		},

		ginkgo.Entry("", "counter", []string{"a", "b"}, "counter|a,b"),
		ginkgo.Entry("", "counter", []string{"b", "a"}, "counter|a,b"),
		ginkgo.Entry("", "counter", []string{"x", "z", "y"}, "counter|x,y,z"),
		ginkgo.Entry("", "counter", []string{"x", "y", "x"}, "counter|x,y"),
		ginkgo.Entry("", "counter", []string{"", "b", "a"}, "counter|a,b"),
		ginkgo.Entry("", "counter", nil, "counter"),
		ginkgo.Entry("", "counter", []string{}, "counter"),
	)

	ginkgo.DescribeTable("should split",
		func(metricID string, xn string, xt []string) {
			name, tags := SplitMetricID(metricID)
			Expect(name).To(Equal(xn))
			Expect(tags).To(Equal(xt))
		},

		ginkgo.Entry("", "counter|a,b", "counter", []string{"a", "b"}),
		ginkgo.Entry("", "|counter|a,b", "|counter", []string{"a", "b"}),
		ginkgo.Entry("", "counter", "counter", nil),
	)
})
