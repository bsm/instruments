package instruments

import (
	"strings"
	"sync"
)

var bufferPool sync.Pool

func pooledBuffer(minCap int) []byte {
	if v := bufferPool.Get(); v != nil {
		if p := v.([]byte); cap(p) <= minCap {
			return p[:0]
		}
	}
	return make([]byte, 0, minCap)
}

// MetricID takes a name and tags and generates a consistent
// metric identifier
func MetricID(name string, tags []string) string {
	if len(tags) == 0 {
		return name
	}

	size := len(name)
	for _, t := range tags {
		if t != "" {
			size += len(t) + 1
		}
	}

	buf := pooledBuffer(size)
	buf = append(buf, name...)
	defer bufferPool.Put(buf)

	for pos, tag := 0, ""; pos < len(tags); pos++ {
		if next := findMinString(tags, tag); next > tag {
			tag = next
		} else {
			break
		}

		if pos == 0 {
			buf = append(buf, '|')
		} else {
			buf = append(buf, ',')
		}
		buf = append(buf, tag...)
	}

	return string(buf)
}

// SplitMetricID takes a metric ID and splits it into
// name and tags.
func SplitMetricID(metricID string) (name string, tags []string) {
	if metricID == "" {
		return "", nil
	}

	pos := strings.LastIndexByte(metricID, '|')
	if pos > 0 && pos < len(metricID)-1 {
		return metricID[:pos], strings.Split(metricID[pos+1:], ",")
	}
	return metricID, nil
}

func findMinString(slice []string, greaterThan string) string {
	min := greaterThan
	for _, s := range slice {
		if s > greaterThan && (min == greaterThan || s < min) {
			min = s
		}
	}
	return min
}
