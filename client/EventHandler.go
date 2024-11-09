package client

import (
	"context"
	"events-to-log/logging"
	"flag"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
)

func Start() {
	client := createKubeClient()

	events, err := client.CoreV1().Events("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for watchedEvent := range events.ResultChan() {
		event, ok := watchedEvent.Object.(*v1.Event)

		if !ok {
			fmt.Printf("Event is not of expected type, skipping")
			continue
		}

		loggableEvent := &logging.LoggableEvent{
			Metadata: logging.Metadata{
				Name:      event.ObjectMeta.Name,
				Namespace: event.ObjectMeta.Namespace,
				UID:       string(event.ObjectMeta.UID),
			},
			Message:   event.Message,
			Timestamp: event.CreationTimestamp.String(),
			Reason:    event.Reason,
			Type:      event.Type,
			Count:     event.Count,
			Reporter:  event.ReportingController,
		}

		logging.Log(loggableEvent)
	}
}

func createKubeClient() *kubernetes.Clientset {
	config, err := parseKubeConfig()

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return client
}

func parseKubeConfig() (*rest.Config, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config, err
}
