package prometheus

import "fmt"

// Client queries the Prometheus HTTP API.
type Client struct {
	endpoint string
}

// NewClient creates a new Prometheus API client.
func NewClient(endpoint string) *Client {
	return &Client{endpoint: endpoint}
}

// QueryLatency returns p50, p95, p99 latency in milliseconds for the given service.
// Returns -1 for each value if the query fails or no data is available.
// Full implementation in LIB-16.
func (c *Client) QueryLatency(service string) (p50, p95, p99 float64, err error) {
	return -1, -1, -1, fmt.Errorf("not implemented yet (LIB-16)")
}
