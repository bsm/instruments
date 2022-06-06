package instruments

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/bsm/ginkgo/v2"
	. "github.com/bsm/gomega"
)

var _ = ginkgo.Describe("Instruments", func() {
	updateInParallel := func(in interface{ Update(float64) }) {
		var wg sync.WaitGroup
		defer wg.Wait()

		update := func() {
			defer wg.Done()

			for i := 0; i < 1000; i++ {
				in.Update(1.0)
			}
		}

		wg.Add(1)
		go update()
		wg.Add(1)
		go update()
	}

	ginkgo.It("should update counters", func() {
		c := NewCounter()
		c.Update(7)
		c.Update(12)
		Expect(c.Snapshot()).To(Equal(19.0))
		Expect(c.Snapshot()).To(Equal(0.0))

		for i := 1; i < 100; i++ {
			c.Update(float64(i))
		}
		Expect(c.Snapshot()).To(Equal(4950.0))
	})

	ginkgo.It("should update counters atomically", func() {
		c := NewCounter()
		updateInParallel(c)
		Expect(c.Snapshot()).To(Equal(2000.0))
	})

	ginkgo.It("should update gauges", func() {
		g := NewGauge()
		g.Update(7)
		g.Update(12)
		Expect(g.Snapshot()).To(Equal(12.0))
	})

	ginkgo.It("should update gauges atomically", func() {
		g := NewGauge()
		updateInParallel(g)
		Expect(g.Snapshot()).To(Equal(1.0))
	})

	ginkgo.It("should update derives", func() {
		d := NewDerive(10)
		d.Update(7)
		time.Sleep(10 * time.Millisecond)
		d.Update(12)
		Expect(d.Snapshot()).To(BeNumerically("~", 200, 50))
	})

	ginkgo.It("should update derives atomically", func() {
		d := NewDerive(10)
		updateInParallel(d)
		Expect(d.Snapshot()).To(BeNumerically("<", 0.0))
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

	ginkgo.It("should update rates atomically", func() {
		r := NewRate()
		updateInParallel(r)
		Expect(r.Snapshot()).To(BeNumerically(">", 0.0))
	})

	ginkgo.It("should update reservoirs", func() {
		r := NewReservoir()
		Expect(r.Snapshot().Count()).To(Equal(0))

		r.Update(1)
		Expect(r.Snapshot().Mean()).To(Equal(1.0))

		r.Update(-10)
		r.Update(23)
		Expect(r.Snapshot().Mean()).To(BeNumerically("~", 4.67, 0.01))
	})

	ginkgo.It("should update reservoirs atomically", func() {
		r := NewReservoir()
		updateInParallel(r)
		Expect(r.Snapshot().Mean()).To(BeNumerically("==", 1.0))
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
