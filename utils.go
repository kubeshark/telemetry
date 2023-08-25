package telemetry

import (
	"context"
	"errors"
	"runtime"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func getCPUUsage() float64 {
	return float64(runtime.NumCPU())
}

func getMemoryUsage() (uint64, uint64) {
	var stat runtime.MemStats
	runtime.ReadMemStats(&stat)
	return stat.Alloc, stat.Sys
}

func getPodInfo(clientSet *kubernetes.Clientset, serviceName string) (internalIP string, namespace string, err error) {
	podClient := clientSet.CoreV1().Pods(metav1.NamespaceDefault)

	var podList *v1.PodList
	podList, err = podClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return
	}

	var hubPod v1.Pod
	for _, pod := range podList.Items {
		if strings.HasPrefix(pod.Name, serviceName) {
			hubPod = pod
		}
	}

	if hubPod.Name == "" {
		err = errors.New("failed to get k8s kubeshark-hub Pod - no Pod with prefix " + serviceName + " was found")
		return
	}

	internalIP = hubPod.Status.PodIP
	namespace = hubPod.ObjectMeta.Namespace

	return
}

func getClusterIP(clientSet *kubernetes.Clientset) (clusterIP string, err error) {
	// TODO: fetch it from svcClient.List and find by ClusterIP type rather than kubernetes name?
	serviceName := "kubernetes"
	var service *v1.Service
	service, err = clientSet.CoreV1().Services("default").Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	clusterIP = service.Spec.ClusterIP
	return
}
