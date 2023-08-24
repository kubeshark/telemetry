package telemetry

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"time"
)

func Run(startTime time.Time, clientSet *kubernetes.Clientset, serviceName string) {
	now := time.Now()
	cpuUsage := getCPUUsage()
	memAlloc, memSys := getMemoryUsage()
	memUsage := float64(memAlloc) / float64(memSys) * 100

	clusterIP, err := getClusterIP(clientSet)
	if err != nil {
		log.Error().Err(err).Msg("Error getting cluster ip")
		return
	}

	_podInternalIp, _podNameSpace, err := getPodInfo(clientSet, serviceName)
	if err != nil {
		log.Error().Err(err).Msg("Error getting cluster ip")
		return
	}

	data := Stats{
		Timestamp:     now,
		TimeFromStart: now.Sub(startTime),
		CPU:           cpuUsage,
		Memory:        memAlloc,
		MemoryUsage:   memUsage,
		ClusterIP:     clusterIP,
		PodIP:         _podInternalIp,
		PodNamespace:  _podNameSpace,
		PodName:       "kubeshark-hub",
	}

	log.Info().Str("timestamp", data.Timestamp.String()).
		Str("time_from_start", data.TimeFromStart.String()).
		Float64("cpu", data.CPU).
		Uint64("memory", data.Memory).
		Float64("mem_usage", data.MemoryUsage).
		Str("node_id", data.NodeID).
		Str("cluster_ip", clusterIP).
		Msg("TELEMETRY")

	if err := emitMetrics(data, serviceName); err != nil {
		log.Error().Err(err).Msg("EMIT METRICS FAILED")
		return
	}
}

func emitMetrics(data Stats, serviceName string) error {
	apiURL := "https://api.kubeshark.co"
	URL := apiURL + "/telemetry/" + serviceName

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
		}
		return fmt.Errorf("unexpected response status: %d with message %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
