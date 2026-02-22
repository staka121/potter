package dashboard

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/staka121/potter/pkg/types"
)

// Row holds the display data for a single service in the dashboard.
type Row struct {
	ServiceName string
	P50Ms       float64 // milliseconds; -1 means N/A (fetch failed)
	P95Ms       float64
	P99Ms       float64
	SLA         *types.LatencyConfig // nil if no SLA defined in contract
}

// SLAStatus represents the overall SLA evaluation result for a row.
type SLAStatus int

const (
	SLAUnknown  SLAStatus = iota // no SLA defined or metrics unavailable
	SLAOk                        // all thresholds satisfied
	SLAViolated                  // at least one threshold exceeded
)

// Evaluate compares actual latencies against the SLA thresholds.
// Returns the status and the first violated percentile label (e.g. "p95").
func (r Row) Evaluate() (SLAStatus, string) {
	if r.SLA == nil || r.P50Ms < 0 {
		return SLAUnknown, ""
	}

	for _, entry := range []struct {
		actual    float64
		threshold string
		label     string
	}{
		{r.P50Ms, r.SLA.P50, "p50"},
		{r.P95Ms, r.SLA.P95, "p95"},
		{r.P99Ms, r.SLA.P99, "p99"},
	} {
		if entry.threshold == "" {
			continue
		}
		ms, err := parseDurationToMs(entry.threshold)
		if err != nil {
			continue
		}
		if entry.actual > ms {
			return SLAViolated, entry.label
		}
	}

	return SLAOk, ""
}

// parseDurationToMs parses a duration string like "100ms", "1s", "500us" into milliseconds.
func parseDurationToMs(s string) (float64, error) {
	s = strings.TrimSpace(s)
	switch {
	case strings.HasSuffix(s, "ms"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "ms"), 64)
		return v, err
	case strings.HasSuffix(s, "us"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "us"), 64)
		return v / 1000, err
	case strings.HasSuffix(s, "s"):
		v, err := strconv.ParseFloat(strings.TrimSuffix(s, "s"), 64)
		return v * 1000, err
	default:
		return 0, fmt.Errorf("unsupported duration unit in %q", s)
	}
}
