package datadog

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// MetricType defines the type of the metric
type MetricType string

// Metric types
const (
	MetricCount MetricType = "count"
	MetricGauge MetricType = "gauge"
	MetricRate  MetricType = "rate"
)

// Metric represents a flushed metric
type Metric struct {
	Name   string     `json:"metric"`
	Type   MetricType `json:"type,omitempty"`
	Points []Point    `json:"points"`
	Host   string     `json:"host,omitempty"`
	Tags   []string   `json:"tags,omitempty"`
}

// Point represents a data point
type Point struct {
	T int64       // the timestamp
	V interface{} // the value, may be a float64/int64/int
}

// MarshalJSON implements json.Marshaler.
func (p Point) MarshalJSON() ([]byte, error) {
	buf := make([]byte, 0, 30)
	buf = append(buf, '[')
	buf = strconv.AppendInt(buf, p.T, 10)
	buf = append(buf, ',')
	switch val := p.V.(type) {
	case float64:
		buf = strconv.AppendFloat(buf, val, 'f', 6, 64)
	case int64:
		buf = strconv.AppendInt(buf, val, 10)
	case int:
		buf = strconv.AppendInt(buf, int64(val), 10)
	}
	return append(buf, ']'), nil
}

// DefaultURL is the default series URL the client sends metric data to
const DefaultURL = "https://app.datadoghq.com/api/v1/series"

// Client abstracts a datadog API connection.
type Client struct {
	apiKey string
	client *http.Client

	// URL is the series URL to push data to.
	// Default: DefaultURL
	URL string

	// Disables zlib payload compression when
	// POSTing data to the API.
	DisableCompression bool
}

// NewClient creates a new API client.
func NewClient(apiKey string) *Client {
	return &Client{
		client: &http.Client{Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			DisableKeepAlives:     true,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}},
		apiKey: apiKey,
		URL:    DefaultURL,
	}
}

// Post delivers a metrics snapshot to datadog
func (c *Client) Post(metrics []Metric) error {
	series := struct {
		Series []Metric `json:"series,omitempty"`
	}{Series: metrics}

	buf := fetchBuffer()
	defer bufferPool.Put(buf)

	var dst io.Writer = buf
	if !c.DisableCompression {
		zlw := fetcZlibWriter(buf)
		defer zlibWriterPool.Put(zlw)
		defer zlw.Close()

		dst = zlw
	}

	if err := json.NewEncoder(dst).Encode(&series); err != nil {
		return err
	}
	if c, ok := dst.(io.Closer); ok {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return c.post(buf.Bytes(), 0)
}

func (c *Client) post(data []byte, retries int) error {
	req, err := http.NewRequest("POST", c.URL+"?api_key="+c.apiKey, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if !c.DisableCompression {
		req.Header.Set("Content-Encoding", "deflate")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	} else if retries <= 3 && resp.StatusCode >= 500 {
		time.Sleep(time.Duration(retries+1) * 200 * time.Second)
		return c.post(data, retries+1)
	} else {
		return fmt.Errorf("datadog: bad API response: %s", resp.Status)
	}
}

// --------------------------------------------------------------------

var (
	bufferPool     sync.Pool
	zlibWriterPool sync.Pool
)

func fetchBuffer() *bytes.Buffer {
	if v := bufferPool.Get(); v != nil {
		b := v.(*bytes.Buffer)
		b.Reset()
		return b
	}
	return new(bytes.Buffer)
}

func fetcZlibWriter(w io.Writer) *zlib.Writer {
	if v := zlibWriterPool.Get(); v != nil {
		z := v.(*zlib.Writer)
		z.Reset(w)
		return z
	}
	return zlib.NewWriter(w)
}
