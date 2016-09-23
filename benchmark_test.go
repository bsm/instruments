package instruments_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/bsm/instruments"
)

func BenchmarkCounter(b *testing.B) {
	c := instruments.NewCounter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Update(int64(i))
		c.Snapshot()
	}
}

func BenchmarkRate(b *testing.B) {
	r := instruments.NewRate()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Update(int64(i))
		r.Snapshot()
	}
}

func BenchmarkReservoir(b *testing.B) {
	r := instruments.NewReservoir(-1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Update(int64(i))
		r.Snapshot()
	}
}

func BenchmarkTimer(b *testing.B) {
	r := instruments.NewTimer(-1)
	s := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Since(s)
		r.Snapshot()
	}
}

func BenchmarkRegistry_Register(b *testing.B) {
	r := instruments.New(time.Minute, "")
	defer r.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Register(fmt.Sprintf("foo.%d", i), nil, instruments.NewRate())
	}
}

func BenchmarkRegistry_reset(b *testing.B) {
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
