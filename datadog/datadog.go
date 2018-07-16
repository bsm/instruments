package datadog

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// DefaultURL is the default series URL the client sends metric data to
const DefaultURL = "https://app.datadoghq.com/api/v1/series"

// retriesCount holds the max number of retries when POST-ing metrics.
const retriesCount = 3

// payload represents HTTP request payload to POST metrics.
type payload struct {
	Series []Metric `json:"series,omitempty"`
}

// Metric represents a flushed metric
type Metric struct {
	Name   string           `json:"metric"`
	Points [][2]interface{} `json:"points"`
	Host   string           `json:"host,omitempty"`
	Tags   []string         `json:"tags,omitempty"`
}

// Client is a DataDog API client for metric flushing.
type Client struct {
	apiKey string
	client *http.Client

	// URL is the series URL to push data to.
	// Default: DefaultURL
	URL string

	bfs, zws sync.Pool

	setHeaders   func(http.Header)
	writePayload func(io.Writer, *payload) error
}

// NewClient creates a new API client.
func NewClient(apiKey string, opts ...Option) *Client {
	o := new(options)
	applyOptions(o, opts...)

	c := &Client{
		apiKey: apiKey,
		client: &http.Client{},
		URL:    DefaultURL,
	}

	if o.noCompression {
		c.setHeaders = func(h http.Header) {
			h.Set("Content-Type", "application/json")
		}
		c.writePayload = writeUncompressedPayload
	} else {
		c.setHeaders = func(h http.Header) {
			h.Set("Content-Type", "application/json")
			h.Set("Content-Encoding", "deflate")
		}
		c.writePayload = func(w io.Writer, p *payload) error {
			zw := c.zWriter(w)
			defer c.zws.Put(zw)
			defer zw.Close()

			if err := writeUncompressedPayload(zw, p); err != nil {
				return err
			}
			return zw.Flush()
		}
	}

	return c
}

// Post delivers a metrics snapshot to datadog.
func (c *Client) Post(metrics []Metric) error {
	p := &payload{Series: metrics}

	buf := c.buffer()
	defer c.bfs.Put(buf)

	if err := c.writePayload(buf, p); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.URL+"?api_key="+c.apiKey, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req.Header)

	for i := 1; i <= retriesCount; i++ {
		req.Body = readCloser{Reader: bytes.NewReader(buf.Bytes())} // make a new reader for each try

		var code int
		code, err = c.post(req)
		if err == nil || code < http.StatusInternalServerError { // only server errors are retried, 4xx are "fatal"
			return err
		}

		time.Sleep(time.Duration(i) * 200 * time.Millisecond)
	}
	return err
}

func (c *Client) post(req *http.Request) (int, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		return resp.StatusCode, nil
	}
	return resp.StatusCode, fmt.Errorf("datadog: bad API response: %s", resp.Status)
}

func (c *Client) buffer() *bytes.Buffer {
	if v := c.bfs.Get(); v != nil {
		b := v.(*bytes.Buffer)
		b.Reset()
		return b
	}
	return new(bytes.Buffer)
}

func (c *Client) zWriter(w io.Writer) *zlib.Writer {
	if v := c.zws.Get(); v != nil {
		z := v.(*zlib.Writer)
		z.Reset(w)
		return z
	}
	return zlib.NewWriter(w)
}

// ----------------------------------------------------------------------------

type readCloser struct {
	io.Reader
}

func (readCloser) Close() error { return nil }

func writeUncompressedPayload(w io.Writer, p *payload) error {
	return json.NewEncoder(w).Encode(p)
}
