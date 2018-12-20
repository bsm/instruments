package logreporter

import (
	"fmt"
	"log"
	"strings"

	"github.com/bsm/instruments"
)

var _ instruments.Reporter = (*Reporter)(nil)

// Logger follows the standard log.Logger API
type Logger interface {
	Println(v ...interface{})
}

// Reporter implements instruments.Reporter and simply logs metrics
type Reporter struct {
	logger  Logger
	metrics []string
}

// New creates a new reporter using a logger.
// Uses log package's default logger if none given
func New(logger Logger) *Reporter {
	return &Reporter{logger: logger}
}

// Prep implements instruments.Reporter
func (r *Reporter) Prep() error { return nil }

// Counting implements instruments.Reporter
func (r *Reporter) Counting(name string, tags []string, val int64) error {
	metric := fmt.Sprintf("%s|%s:val=%d", name, strings.Join(tags, ","), val)
	r.metrics = append(r.metrics, metric)
	return nil
}

// Discrete implements instruments.Reporter
func (r *Reporter) Discrete(name string, tags []string, val float64) error {
	metric := fmt.Sprintf("%s|%s:val=%v", name, strings.Join(tags, ","), val)
	r.metrics = append(r.metrics, metric)
	return nil
}

// Sample implements instruments.Reporter
func (r *Reporter) Sample(name string, tags []string, dist instruments.Distribution) error {
	metric := fmt.Sprintf("%s|%s:p95=%v", name, strings.Join(tags, ","), dist.Quantile(0.95))
	r.metrics = append(r.metrics, metric)
	return nil
}

// Flush implements instruments.Reporter
func (r *Reporter) Flush() error {
	r.log(strings.Join(r.metrics, " "))
	r.metrics = r.metrics[:0]
	return nil
}

func (r *Reporter) log(v ...interface{}) {
	if r.logger != nil {
		r.logger.Println(v...)
	} else {
		log.Println(v...)
	}
}
