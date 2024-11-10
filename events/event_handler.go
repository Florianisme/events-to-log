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
)

type Watcher struct {
	events    watch.Interface
	persister *persistence.TimestampPersister
}

func Init() *Watcher {
	kubeClient := client.CreateKubeClient()

	persister := persistence.Init(kubeClient)
	events := startEventWatch(kubeClient)

	return &Watcher{
		events:    events,
		persister: persister,
	}
}

func (s *Watcher) StartWatching() {
	for watchedEvent := range s.events.ResultChan() {
		event, ok := watchedEvent.Object.(*v1.Event)

		if !ok {
			fmt.Printf("event of type %s can not be mapped, skipping\n", event.Type)
			continue
		}

		if eventAlreadyProcessed(event, s) {
			fmt.Printf("event has already been processed, skipping (at %s)\n", event.CreationTimestamp.String())
			continue
		}

		loggableEvent := mapLoggableEvent(event)
		logging.Log(loggableEvent)

		s.persister.UpdateCurrentTimestamp(event.CreationTimestamp.Time)
	}
}

func eventAlreadyProcessed(event *v1.Event, s *Watcher) bool {
	return event.CreationTimestamp.Time.Equal(s.persister.GetCurrentTimestamp()) ||
		event.CreationTimestamp.Time.Before(s.persister.GetCurrentTimestamp())
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
	events, err := client.CoreV1().Events("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	return events
}

func (s *Watcher) StopWatching() {
	s.events.Stop()
	s.persister.Flush()
}
