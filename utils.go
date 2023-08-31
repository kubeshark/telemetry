package telemetry

import (
	"os"
	"runtime"
)

func getCPUUsage() float64 {
	return float64(runtime.NumCPU())
}

func getMemoryUsage() (uint64, uint64) {
	var stat runtime.MemStats
	runtime.ReadMemStats(&stat)
	return stat.Alloc, stat.Sys
}

func getHostname() (string, error) {
	content, err := os.ReadFile("/etc/hostname")
	if err != nil {
		return "", err
	}

	return string(content), nil
}
