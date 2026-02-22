package dashboard

import (
	"fmt"
	"time"
)

// Render clears the terminal and draws the dashboard table.
// Full implementation in LIB-17.
func Render(tsuboName, prometheusEndpoint string, rows []Row, updatedAt time.Time) {
	// Placeholder: simple text output until LIB-17 implements full CUI rendering
	fmt.Printf("\033[H\033[2J") // clear screen
	fmt.Printf("Potter Monitor Dashboard — %s\n", tsuboName)
	fmt.Printf("prometheus: %s   updated: %s\n\n", prometheusEndpoint, updatedAt.Format("15:04:05"))
	fmt.Printf("%-20s %8s %8s %8s %s\n", "SERVICE", "p50", "p95", "p99", "SLA")
	fmt.Println("--------------------------------------------------------------")
	for _, row := range rows {
		p50 := fmtMs(row.P50Ms)
		p95 := fmtMs(row.P95Ms)
		p99 := fmtMs(row.P99Ms)
		status, violated := row.Evaluate()
		sla := fmtSLA(status, violated)
		fmt.Printf("%-20s %8s %8s %8s %s\n", row.ServiceName, p50, p95, p99, sla)
	}
}

func fmtMs(v float64) string {
	if v < 0 {
		return "N/A"
	}
	return fmt.Sprintf("%.0fms", v)
}

func fmtSLA(status SLAStatus, violated string) string {
	switch status {
	case SLAOk:
		return "✓"
	case SLAViolated:
		return "✗ " + violated
	default:
		return "-"
	}
}
