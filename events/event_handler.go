package events

import (
	"context"
	"events-to-log/client"
	"events-to-log/logging"
	"events-to-log/persistence"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

func StartWatching() {
	kubeClient := client.CreateKubeClient()
	persistence.Init(kubeClient)
	defer persistence.Shutdown(kubeClient)

	events := startEventWatch(kubeClient)

	for watchedEvent := range events.ResultChan() {
		event, ok := watchedEvent.Object.(*v1.Event)

		if !ok {
			status := watchedEvent.Object.(*metav1.Status)
			if status.Code == http.StatusGone {
				// We have to reset the current resource version because it's too old
				fmt.Println("currently stored ResourceVersion is too old. " +
					"It will be deleted and the application will restart")
				persistence.UpdateCurrentResourceVersion("")
			}
			continue
		}

		loggableEvent := mapLoggableEvent(event)

		logging.Log(loggableEvent)
		persistence.UpdateCurrentResourceVersion(event.ResourceVersion)
	}
}

func mapLoggableEvent(event *v1.Event) *logging.LoggableEvent {
	loggableEvent := &logging.LoggableEvent{
		Metadata: logging.Metadata{
			Name:            event.ObjectMeta.Name,
			Namespace:       event.ObjectMeta.Namespace,
			UID:             string(event.ObjectMeta.UID),
			ResourceVersion: event.ObjectMeta.ResourceVersion,
		},
		Message:   event.Message,
		Timestamp: event.CreationTimestamp.String(),
		Reason:    event.Reason,
		Type:      event.Type,
		Count:     event.Count,
		Reporter:  event.ReportingController,
	}
	return loggableEvent
}

func startEventWatch(client *kubernetes.Clientset) watch.Interface {
	options := metav1.ListOptions{
		ResourceVersion: persistence.GetCurrentResourceVersion(),
	}

	events, err := client.CoreV1().Events("").Watch(context.TODO(), options)
	if err != nil {
		panic(err)
	}
	return events
}
