package telemetry

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"
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

func Run(startTime time.Time, clientSet *kubernetes.Clientset, serviceName string) (stats *Stats, err error) {
	now := time.Now()
	cpuUsage := getCPUUsage()
	memAlloc, memSys := getMemoryUsage()
	memUsage := float64(memAlloc) / float64(memSys) * 100

	var clusterIP string
	clusterIP, err = getClusterIP(clientSet)
	if err != nil {
		return
	}

	var podInternalIp, podNamespace string
	podInternalIp, podNamespace, err = getPodInfo(clientSet, serviceName)
	if err != nil {
		return
	}

	stats = &Stats{
		Timestamp:     now,
		TimeFromStart: now.Sub(startTime),
		CPU:           cpuUsage,
		Memory:        memAlloc,
		MemoryUsage:   memUsage,
		ClusterIP:     clusterIP,
		PodIP:         podInternalIp,
		PodNamespace:  podNamespace,
		PodName:       "kubeshark-hub",
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
