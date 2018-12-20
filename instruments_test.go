package instruments

import (
	"math/rand"
	"testing"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Instruments", func() {

	DescribeTable("Reservoir",
		func(vv []float64, x float64) {
			i := NewReservoir()
			for _, v := range vv {
				i.Update(v)
			}
			Expect(i.Snapshot().Mean()).To(BeNumerically("~", x, 0.1))
		},
		Entry("single", []float64{1}, 1.0),
		Entry("a few", []float64{1, -10, 23}, 4.7),
	)

	ginkgo.It("should update counters", func() {
		c := NewCounter()
		c.Update(7)
		c.Update(12)
		Expect(c.Snapshot()).To(Equal(int64(19)))
		Expect(c.Snapshot()).To(Equal(int64(0)))

		for i := 1; i < 100; i++ {
			c.Update(int64(i))
		}
		Expect(c.Snapshot()).To(Equal(int64(4950)))
	})

	ginkgo.It("should update gauges", func() {
		g := NewGauge()
		g.Update(7)
		g.Update(12)
		Expect(g.Snapshot()).To(Equal(12.0))
	})

	ginkgo.It("should update derives", func() {
		d := NewDerive(10)
		d.Update(7)
		time.Sleep(10 * time.Millisecond)
		d.Update(12)
		Expect(d.Snapshot()).To(BeNumerically("~", 200, 50))
	})

	ginkgo.It("should update rates", func() {
		r := NewRate()
		Expect(r.Snapshot()).To(Equal(0.0))

		r.Update(100)
		time.Sleep(10 * time.Millisecond)
		Expect(r.Snapshot()).To(BeNumerically("~", 10000, 2500))

		r.Update(100)
		time.Sleep(time.Millisecond)
		Expect(r.Snapshot()).To(BeNumerically("~", 100000, 40000))
	})

	ginkgo.It("should update timers", func() {
		t := NewTimer()
		for i := 0; i < 100; i++ {
			t.Update(time.Millisecond * time.Duration(i))
		}
		s := t.Snapshot()
		Expect(s.Mean()).To(BeNumerically("~", 49.5, 0.01))
		Expect(s.Quantile(0.75)).To(BeNumerically("~", 74.5, 0.01))
	})

})

// --------------------------------------------------------------------

func init() {
	rand.Seed(5)
}

func TestSuite(t *testing.T) {
	RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "instruments")
}

// --------------------------------------------------------------------

// test exports

func (r *Registry) GetInstruments() map[string]interface{}            { return r.instruments }
func (r *Registry) SetInstruments(instruments map[string]interface{}) { r.instruments = instruments }
func (r *Registry) Reset() int                                        { return len(r.reset()) }

func ReleaseDistribution(d Distribution) {
	releaseDistribution(d)
}
