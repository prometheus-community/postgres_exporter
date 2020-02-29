package servers

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type cachedMetrics struct {
	metrics    []prometheus.Metric
	lastScrape time.Time
}