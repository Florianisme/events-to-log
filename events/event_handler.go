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
	"k8s.io/client-go/tools/cache"
	watch2 "k8s.io/client-go/tools/watch"
	"time"
)

type Watcher struct {
	watcher   *watch2.RetryWatcher
	persister *persistence.TimestampPersister
	logger    *logging.Logger
}

func Init() *Watcher {
	kubeClient := client.CreateKubeClient()

	logger := logging.Init()
	persister := persistence.Init(kubeClient, logger)
	retryWatcher := startEventWatch(kubeClient)

	return &Watcher{
		watcher:   retryWatcher,
		persister: persister,
		logger:    logger,
	}
}

func (s *Watcher) StartWatching() {
	for watchedEvent := range (*s.watcher).ResultChan() {
		event, ok := watchedEvent.Object.(*v1.Event)

		if !ok {
			if _, statusOk := watchedEvent.Object.(*metav1.Status); statusOk {
				// event channel was probably closed, we should safely return here
				return
			}

			s.logger.Logger.Debug().Msgf("event of type %s can not be mapped, skipping", event.Type)
			continue
		}

		timestamp := getComparableTimestamp(event)
		if eventAlreadyProcessed(timestamp, s) {
			s.logger.Logger.Debug().Msgf("event %s in namespace %s has already been processed, skipping (at %s)",
				event.ObjectMeta.Name, event.ObjectMeta.Namespace, timestamp.UTC().String())
			continue
		}

		loggableEvent := mapLoggableEvent(event)
		s.logger.Log(loggableEvent)

		s.persister.UpdateCurrentTimestamp(timestamp.Time)
	}
	s.logger.Logger.Debug().Msg("end of channel reached, no more watcher will be processed")
}

func eventAlreadyProcessed(timestamp metav1.Time, s *Watcher) bool {
	// We allow up to 3 seconds of buffer here. In case loads of watcher are being created at once, we might miss them otherwise.
	// The chance of processing an event twice after restart is relatively low compared to missing one otherwise.
	eventAge := s.persister.GetCurrentTimestamp().Sub(timestamp.Time)
	return eventAge > (3 * time.Second)
}

func mapLoggableEvent(event *v1.Event) *logging.LoggableEvent {
	loggableEvent := &logging.LoggableEvent{
		Metadata: logging.Metadata{
			Name:            event.ObjectMeta.Name,
			Namespace:       event.ObjectMeta.Namespace,
			UID:             string(event.ObjectMeta.UID),
			ResourceVersion: event.ObjectMeta.ResourceVersion,
		},
		Message:            event.Message,
		CreationTimestamp:  mapTimestamp(event.CreationTimestamp),
		FirstSeenTimestamp: mapTimestamp(event.FirstTimestamp),
		LastSeenTimestamp:  mapTimestamp(event.LastTimestamp),
		Reason:             event.Reason,
		Type:               event.Type,
		Count:              event.Count,
		Reporter:           event.ReportingController,
	}
	return loggableEvent
}

func mapTimestamp(timestamp metav1.Time) string {
	if timestamp.IsZero() {
		return ""
	}

	return timestamp.Local().String()
}

// getComparableTimestamp returns a comparable timestamp for this event.
// As some timestamps are optional (but higher in the order of importance), we traverse them in the following
// order and return the first non-empty timestamp:
// LastTimestamp (optional) or else CreationTimestamp (always set)
func getComparableTimestamp(event *v1.Event) metav1.Time {
	var timestamp = event.LastTimestamp

	if !timestamp.IsZero() {
		return timestamp
	}

	return event.CreationTimestamp
}

func startEventWatch(client *kubernetes.Clientset) *watch2.RetryWatcher {
	watchFunc := func(options metav1.ListOptions) (watch.Interface, error) {
		// Return watcher with 30m timeout
		timeoutSeconds := int64(30 * 60 * 60)
		return client.CoreV1().Events("").Watch(context.TODO(), metav1.ListOptions{TimeoutSeconds: &timeoutSeconds})
	}

	watcher, err := watch2.NewRetryWatcher("1", &cache.ListWatch{WatchFunc: watchFunc})
	if err != nil {
		panic(err)
	}
	return watcher
}

func (s *Watcher) StopWatching() {
	(*s.watcher).Stop()
	s.persister.Flush()
}
