package instruments_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/bsm/instruments"
)

func BenchmarkCounter(b *testing.B) {
	c := instruments.NewCounter()
	benchmarkInstrument(b, func(i int) {
		c.Update(float64(i))
		if i%10 == 0 {
			c.Snapshot()
		}
	})
}

func BenchmarkRate(b *testing.B) {
	r := instruments.NewRate()
	benchmarkInstrument(b, func(i int) {
		r.Update(float64(i))
		if i%10 == 0 {
			r.Snapshot()
		}
	})
}

func BenchmarkGauge(b *testing.B) {
	g := instruments.NewGauge()
	benchmarkInstrument(b, func(i int) {
		g.Update(float64(i))
		if i%10 == 0 {
			g.Snapshot()
		}
	})
}

func BenchmarkDerive(b *testing.B) {
	d := instruments.NewDerive(10)
	benchmarkInstrument(b, func(i int) {
		d.Update(float64(i))
		if i%10 == 0 {
			d.Snapshot()
		}
	})
}

func BenchmarkReservoir(b *testing.B) {
	r := instruments.NewReservoir()
	benchmarkInstrument(b, func(i int) {
		r.Update(float64(i))
		if i%10 == 0 {
			instruments.ReleaseDistribution(r.Snapshot())
		}
	})
}

func BenchmarkTimer(b *testing.B) {
	r := instruments.NewTimer()
	s := time.Now()
	benchmarkInstrument(b, func(i int) {
		r.Since(s)
		if i%10 == 0 {
			instruments.ReleaseDistribution(r.Snapshot())
		}
	})
}

func BenchmarkRegistry_Register(b *testing.B) {
	r := instruments.New(time.Minute, "")
	defer r.Close()

	counter := instruments.NewCounter()
	tags := []string{"k1:v1", "k2:v2"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Register(fmt.Sprintf("foo.%d", i), tags, counter)
	}
}

func BenchmarkRegistry_Reset(b *testing.B) {
	n := 10000
	s := instruments.New(time.Minute, "")
	defer s.Close()
	for i := 0; i < n; i++ {
		s.Register(fmt.Sprintf("foo.%d", i), nil, instruments.NewRate())
	}
	m := s.GetInstruments()
	r := instruments.New(time.Minute, "")
	defer r.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.SetInstruments(m)
		if size := r.Reset(); size != n {
			b.Fatal("snapshot returned unexpected size:", size)
		}
	}
}

func BenchmarkMetricID(b *testing.B) {
	tags := []string{"foo", "bar", "baz", "doh"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		instruments.MetricID("metric", tags)
	}
}

func benchmarkInstrument(b *testing.B, cb func(int)) {
	b.Helper()

	b.Run("serial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cb(i)
		}
	})
	b.Run("parallel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				cb(i)
			}
		})
	})
}
