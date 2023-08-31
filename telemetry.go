package telemetry

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

var cloudApiURL = "https://api.kubeshark.co"

const (
	ENV_CLOUD_API_URL = "KUBESHARK_CLOUD_API_URL"
)

func init() {
	envCloudApiURL := os.Getenv(ENV_CLOUD_API_URL)
	if envCloudApiURL != "" {
		cloudApiURL = envCloudApiURL
	}
}

const (
	ENV_TELEMETRY_DISABLED             = "TELEMETRY_DISABLED"
	ENV_TELEMETRY_INTERVAL_SECONDS     = "TELEMETRY_INTERVAL_SECONDS"
	DEFAULT_TELEMETRY_INTERVAL_SECONDS = 60
)

func Start(serviceName string) {
	telemetryDisabled := os.Getenv(ENV_TELEMETRY_DISABLED)
	log.Debug().Str(ENV_TELEMETRY_DISABLED, telemetryDisabled).Msg("Environment variable:")

	if telemetryDisabled == "true" {
		return
	}

	telemetryIntervalSecondsEnv := os.Getenv(ENV_TELEMETRY_INTERVAL_SECONDS)
	log.Debug().Str(ENV_TELEMETRY_INTERVAL_SECONDS, telemetryIntervalSecondsEnv).Msg("Environment variable:")
	telemetryIntervalSeconds, err := strconv.Atoi(telemetryIntervalSecondsEnv)
	if err != nil {
		telemetryIntervalSeconds = DEFAULT_TELEMETRY_INTERVAL_SECONDS
	}

	startTime := time.Now()

	ticker := time.NewTicker(time.Second * time.Duration(telemetryIntervalSeconds))
	defer ticker.Stop()
	for range ticker.C {
		stats, err := Run(startTime, serviceName)
		if err != nil {
			log.Warn().Err(err).Msg("Telemetry")
		} else {
			log.Debug().Interface("stats", stats).Msg("Telemetry")
		}
	}
}

func Run(startTime time.Time, serviceName string) (stats *Stats, err error) {
	now := time.Now()
	cpuUsage := getCPUUsage()
	memAlloc, memSys := getMemoryUsage()
	memUsage := float64(memAlloc) / float64(memSys) * 100

	stats = &Stats{
		Timestamp:     now,
		TimeFromStart: now.Sub(startTime),
		CPU:           cpuUsage,
		Memory:        memAlloc,
		MemoryUsage:   memUsage,
	}

	err = emitMetrics(stats, serviceName)
	return
}

func emitMetrics(data *Stats, serviceName string) (err error) {
	endpointURL := fmt.Sprintf("%s/telemetry/%s", cloudApiURL, serviceName)

	var payload []byte
	payload, err = json.Marshal(data)
	if err != nil {
		return
	}

	var req *http.Request
	req, err = http.NewRequest("POST", endpointURL, bytes.NewBuffer(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	var resp *http.Response
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var bodyBytes []byte
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("unexpected response status: %d error: %s", resp.StatusCode, err.Error())
		} else {
			err = fmt.Errorf("unexpected response status: %d message: %s", resp.StatusCode, string(bodyBytes))
		}
	}

	return
}
