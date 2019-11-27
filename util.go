package instruments

import (
	"math"
	"strings"
)

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

	buf := make([]byte, len(name), size)
	copy(buf, name)

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

// SplitMetricID takes a metric ID ans splits it into
// name and tags
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

// normalizeFloat64 converts NaN/Inf value to 0 or returns it as is otherwise.
func normalizeFloat64(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
