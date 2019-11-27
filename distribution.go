package instruments

import (
	"sync"

	"github.com/bsm/histogram"
)

// Distribution is returned by Sample snapshots
type Distribution interface {
	// Count returns the number of observations
	Count() int
	// Min returns the minimum observed value
	Min() float64
	// Max returns the maximum observed value
	Max() float64
	// Sum returns the sum
	Sum() float64
	// Mean returns the mean
	Mean() float64
	// Quantile returns the quantile for a given q (0..1)
	Quantile(q float64) float64
	// Variance returns the variance
	Variance() float64
}

// NormalizedDistribution converts any NaN/Inf values to zeros.
func NormalizedDistribution(d Distribution) Distribution {
	return normalizedDistribution{d}
}

type normalizedDistribution struct {
	d Distribution
}

func (d normalizedDistribution) Count() int                 { return d.d.Count() }
func (d normalizedDistribution) Min() float64               { return normalizeFloat64(d.d.Min()) }
func (d normalizedDistribution) Max() float64               { return normalizeFloat64(d.d.Max()) }
func (d normalizedDistribution) Sum() float64               { return normalizeFloat64(d.d.Sum()) }
func (d normalizedDistribution) Mean() float64              { return normalizeFloat64(d.d.Mean()) }
func (d normalizedDistribution) Quantile(q float64) float64 { return normalizeFloat64(d.d.Quantile(q)) }
func (d normalizedDistribution) Variance() float64          { return normalizeFloat64(d.d.Variance()) }

// --------------------------------------------------------------------

const defaultHistogramSize = 20

var histogramPool sync.Pool

func newHistogram(sz int) (h *histogram.Histogram) {
	if v := histogramPool.Get(); v != nil {
		h = v.(*histogram.Histogram)
	} else {
		h = new(histogram.Histogram)
	}
	h.Reset(sz)
	return
}

func releaseHistogram(h *histogram.Histogram) {
	histogramPool.Put(h)
}

func releaseDistribution(d Distribution) {
	if h, ok := d.(*histogram.Histogram); ok {
		releaseHistogram(h)
	}
}
