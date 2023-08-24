package telemetry

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"runtime"
	"strings"
)

func getCPUUsage() float64 {
	return float64(runtime.NumCPU())
}

func getMemoryUsage() (uint64, uint64) {
	var stat runtime.MemStats
	runtime.ReadMemStats(&stat)
	return stat.Alloc, stat.Sys
}

func getPodInfo(clientSet *kubernetes.Clientset, serviceName string) (string, string, error) {
	podClient := clientSet.CoreV1().Pods(metav1.NamespaceDefault)

	podList, err := podClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Error listing pods")
	}

	var hubPod v1.Pod
	for _, pod := range podList.Items {
		if strings.HasPrefix(pod.Name, serviceName) {
			log.Info().Interface("pod", pod).Msg("Hub Pod")
			hubPod = pod
		}
	}

	if hubPod.Name == "" {
		return "", "", errors.New("failed to get k8s kubeshark-hub Pod - no Pod with prefix " + serviceName + " was found")
	}

	return hubPod.Status.PodIP, hubPod.ObjectMeta.Namespace, nil
}

func getClusterIP(clientSet *kubernetes.Clientset) (string, error) {
	svcClient := clientSet.CoreV1().Services("default")

	// todo: fetch it from svcClient.List and find by ClusterIP type rather than kubernetes name?
	serviceName := "kubernetes"
	service, err := svcClient.Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return service.Spec.ClusterIP, nil
}
