package telemetry

import "time"

type Stats struct {
	Timestamp     time.Time     `json:"timestamp"`
	TimeFromStart time.Duration `json:"timeFromStart"`
	CPUUsage      float64       `json:"cpuUsage"`
	CPUNum        float64       `json:"cpuNum"`
	MemoryAlloc   uint64        `json:"memoryAlloc"`
	MemoryUsage   float64       `json:"memoryUsage"`
	Hostname      string        `json:"hostname"`
}
