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

type Watcher struct {
	events    watch.Interface
	persister *persistence.ResourceVersionPersister
}

func Init() *Watcher {
	kubeClient := client.CreateKubeClient()

	persister := persistence.Init(kubeClient)
	events := startEventWatch(kubeClient, persister)

	return &Watcher{
		events:    events,
		persister: persister,
	}
}

func (s *Watcher) StartWatching() {
	for watchedEvent := range s.events.ResultChan() {
		event, ok := watchedEvent.Object.(*v1.Event)

		if !ok {
			status := watchedEvent.Object.(*metav1.Status)
			if status.Code == http.StatusGone {
				// We have to reset the current resource version because it's too old
				fmt.Println("currently stored ResourceVersion is too old. " +
					"It will be deleted and the application will restart")
				s.persister.UpdateCurrentResourceVersion("")
			}
			continue
		}

		loggableEvent := mapLoggableEvent(event)
		logging.Log(loggableEvent)

		s.persister.UpdateCurrentResourceVersion(event.ResourceVersion)
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

func startEventWatch(client *kubernetes.Clientset, persister *persistence.ResourceVersionPersister) watch.Interface {

	yes := new(bool)
	*yes = true

	options := metav1.ListOptions{
		SendInitialEvents:    yes,
		Watch:                true,
		ResourceVersionMatch: metav1.ResourceVersionMatchNotOlderThan,
		ResourceVersion:      persister.GetCurrentResourceVersion(),
	}

	events, err := client.CoreV1().Events("").Watch(context.TODO(), options)
	if err != nil {
		panic(err)
	}
	return events
}

func (s *Watcher) StopWatching() {
	s.events.Stop()
	s.persister.Flush()
}
