package persistence

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

var namespace = "event-logging"
var resourceVersionConfigMapName = "resource-version"

var currentResourceVersion string

func Init(client *kubernetes.Clientset) {
	configMap, err := getConfigMap(client)
	if err != nil {
		configMap = createResourceVersionConfigMap(client)
	}

	resolvedResourceVersion := configMap.Data["current"]
	if len(resolvedResourceVersion) == 0 {
		fmt.Printf("no resource version found, starting to watch events from the start\n")
		currentResourceVersion = ""
	} else {
		currentResourceVersion = resolvedResourceVersion
	}

	ticker := time.NewTicker(5 * time.Second)
	go updateConfigMapOnTick(ticker, client)
}

func updateConfigMapOnTick(ticker *time.Ticker, client *kubernetes.Clientset) {
	for range ticker.C {
		configMap, err := getConfigMap(client)
		if err != nil {
			panic(err)
		}

		configMap.Data["current"] = currentResourceVersion

		_, err = client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
		if err != nil {
			panic(err)
		}
	}
}

func createResourceVersionConfigMap(client *kubernetes.Clientset) *v1.ConfigMap {
	configMap, err := client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceVersionConfigMapName,
			Namespace: namespace,
		},
		Data: map[string]string{"current": ""},
	}, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}

	return configMap
}

func UpdateCurrentResourceVersion(updatedResourceVersion string) {
	currentResourceVersion = updatedResourceVersion
}

func GetCurrentResourceVersion() string {
	return currentResourceVersion
}

func getConfigMap(client *kubernetes.Clientset) (*v1.ConfigMap, error) {
	return client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), resourceVersionConfigMapName, metav1.GetOptions{})
}
