package events

import (
	"context"
	"events-to-log/client"
	"events-to-log/logging"
	"events-to-log/persistence"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"time"
)

type Watcher struct {
	events    *watch.Interface
	persister *persistence.TimestampPersister
	logger    *logging.Logger
}

func Init() *Watcher {
	kubeClient := client.CreateKubeClient()

	logger := logging.Init()
	persister := persistence.Init(kubeClient, logger)
	events := startEventWatch(kubeClient)

	return &Watcher{
		events:    events,
		persister: persister,
		logger:    logger,
	}
}

func (s *Watcher) StartWatching() {
	for watchedEvent := range (*s.events).ResultChan() {
		event, ok := watchedEvent.Object.(*v1.Event)

		if !ok {
			if _, statusOk := watchedEvent.Object.(*metav1.Status); statusOk {
				// event channel was probably closed, we should safely return here
				return
			}

			s.logger.Logger.Debug().Msgf("event of type %s can not be mapped, skipping", event.Type)
			continue
		}

		if eventAlreadyProcessed(event, s) {
			s.logger.Logger.Debug().Msgf("event has already been processed, skipping (at %s)", event.CreationTimestamp.String())
			continue
		}

		loggableEvent := mapLoggableEvent(event)
		s.logger.Log(loggableEvent)

		s.persister.UpdateCurrentTimestamp(event.CreationTimestamp.Time)
	}
}

func eventAlreadyProcessed(event *v1.Event, s *Watcher) bool {
	// We allow up to 3 seconds of buffer here. In case loads of events are being created at once, we might miss them otherwise.
	// The chance of processing an event twice after restart is relatively low compared to missing one otherwise.
	return s.persister.GetCurrentTimestamp().Sub(event.CreationTimestamp.Time) > (3 * time.Second)
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

func startEventWatch(client *kubernetes.Clientset) *watch.Interface {
	events, err := client.CoreV1().Events("").Watch(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	return &events
}

func (s *Watcher) StopWatching() {
	(*s.events).Stop()
	s.persister.Flush()
}
