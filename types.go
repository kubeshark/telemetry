package telemetry

import "time"

type Stats struct {
	Timestamp     time.Time
	TimeFromStart time.Duration
	CPU           float64
	Memory        uint64
	MemoryUsage   float64
	ClusterIP     string
	PodIP         string
	PodNamespace  string
	PodName       string
	NodeID        string
}
