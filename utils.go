package telemetry

import (
	"context"
	"errors"
	"runtime"

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

func getPodInfo(clientSet *kubernetes.Clientset, serviceName string) (name string, internalIP string, namespace string, err error) {
	podList, err := clientSet.CoreV1().Pods(metav1.NamespaceDefault).List(
		context.TODO(),
		metav1.ListOptions{
			Watch:         true,
			LabelSelector: "app=" + serviceName,
		},
	)
	if err != nil {
		return
	}

	if len(podList.Items) < 1 {
		err = errors.New("failed to get k8s Pod - no Pod with prefix " + serviceName + " was found")
		return
	}

	var pod v1.Pod
	pod = podList.Items[0]

	name = pod.Name
	internalIP = pod.Status.PodIP
	namespace = pod.ObjectMeta.Namespace

	return
}

func getClusterIP(clientSet *kubernetes.Clientset) (clusterIP string, err error) {
	// TODO: fetch it from svcClient.List and find by ClusterIP type rather than kubernetes name?
	serviceName := "kubernetes"
	var service *v1.Service
	service, err = clientSet.CoreV1().Services("default").Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		return
	}

	clusterIP = service.Spec.ClusterIP
	return
}
