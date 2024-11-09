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

type ResourceVersionPersister struct {
	client                 *kubernetes.Clientset
	ticker                 *time.Ticker
	currentResourceVersion string
}

func Init(client *kubernetes.Clientset) *ResourceVersionPersister {
	configMap, err := getConfigMap(client)
	var currentResourceVersion string

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

	persister := &ResourceVersionPersister{
		client:                 client,
		ticker:                 ticker,
		currentResourceVersion: currentResourceVersion,
	}

	go persister.updateConfigMap()
	return persister
}

func (s *ResourceVersionPersister) updateConfigMap() {
	for range s.ticker.C {
		configMap, err := getConfigMap(s.client)
		if err != nil {
			panic(err)
		}

		configMap.Data["current"] = s.currentResourceVersion

		_, err = s.client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
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

func (s *ResourceVersionPersister) UpdateCurrentResourceVersion(updatedResourceVersion string) {
	s.currentResourceVersion = updatedResourceVersion
}

func (s *ResourceVersionPersister) GetCurrentResourceVersion() string {
	return s.currentResourceVersion
}

func getConfigMap(client *kubernetes.Clientset) (*v1.ConfigMap, error) {
	return client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), resourceVersionConfigMapName, metav1.GetOptions{})
}
func (s *ResourceVersionPersister) Flush() {
	fmt.Printf("flushing last received ResourceVersion %d of event to ConfigMap\n", s.currentResourceVersion)
	s.updateConfigMap()
}
