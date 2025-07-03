package persistence

import (
	"context"
	"events-to-log/logging"
	"os"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var namespace string
var timestampConfigMapName = "timestamp"

type TimestampPersister struct {
	client           *kubernetes.Clientset
	ticker           *time.Ticker
	currentTimestamp time.Time
	logger           *logging.Logger
	done             chan bool
}

func Init(client *kubernetes.Clientset, logger *logging.Logger) *TimestampPersister {
	namespace = getConfiguredNamespace()
	configMap, err := getConfigMap(client)
	var currentTimestamp time.Time

	if err != nil {
		configMap = createTimestampConfigMap(client)
	}

	resolvedTimestamp := configMap.Data["current"]
	if len(resolvedTimestamp) == 0 {
		logger.Logger.Debug().Msg("no timestamp found, starting to watch events from the start")
		currentTimestamp = time.UnixMilli(0)
	} else {
		convertedTimestamp, err := strconv.Atoi(resolvedTimestamp)
		if err != nil {
			logger.Logger.Debug().Msg("malformed timestamp found, starting from the start")
			currentTimestamp = time.UnixMilli(0)
		} else {
			currentTimestamp = time.UnixMilli(int64(convertedTimestamp)).UTC()
			logger.Logger.Debug().Msgf("timestamp to pick up at found, starting event logging at %s", currentTimestamp.String())
		}
	}

	ticker := time.NewTicker(5 * time.Second)
	done := make(chan bool)

	persister := &TimestampPersister{
		client:           client,
		ticker:           ticker,
		currentTimestamp: currentTimestamp,
		logger:           logger,
		done:             done,
	}

	go persister.runScheduledUpdates()
	return persister
}

// getConfiguredNamespace returns the namespace in which the ConfigMap should be updated.
// It returns the value in the following priority order:
// Environment variable "POD_NAMESPACE", Environment variable "NAMESPACE" or else "event-logging" by default
func getConfiguredNamespace() string {
	if podNamespace, isSet := os.LookupEnv("POD_NAMESPACE"); isSet {
		return podNamespace
	}
	if namespace, isSet := os.LookupEnv("NAMESPACE"); isSet {
		return namespace
	}

	return "event-logging"
}

func (s *TimestampPersister) runScheduledUpdates() {
	for {
		select {
		case <-s.done:
			return
		case <-s.ticker.C:
			s.UpdateCurrentTimestamp(time.Now())
			s.updateConfigMap()
		}
	}
}

func (s *TimestampPersister) updateConfigMap() {
	configMap, err := getConfigMap(s.client)
	if err != nil {
		panic(err)
	}

	currentlySavedTimestamp := configMap.Data["current"]
	updatedTimestamp := strconv.FormatInt(s.currentTimestamp.UnixMilli(), 10)

	if currentlySavedTimestamp == updatedTimestamp {
		// nothing to do
		return
	}

	configMap.Data["current"] = updatedTimestamp
	s.logger.Logger.Debug().Msgf("updating current timestamp to %s (unix millis: %d",
		s.currentTimestamp.String(), s.currentTimestamp.UnixMilli())

	_, err = s.client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		panic(err)
	}
}

func createTimestampConfigMap(client *kubernetes.Clientset) *v1.ConfigMap {
	configMap, err := client.CoreV1().ConfigMaps(namespace).Create(context.TODO(), &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      timestampConfigMapName,
			Namespace: namespace,
		},
		Data: map[string]string{"current": "0"},
	}, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}

	return configMap
}

func (s *TimestampPersister) UpdateCurrentTimestamp(updatedTimestamp time.Time) {
	updatedTimestamp = updatedTimestamp.UTC()

	if updatedTimestamp.Before(s.currentTimestamp) {
		s.logger.Logger.Debug().Msgf("not updating timestamp because it's (%s) older than the current one (%s)",
			updatedTimestamp.String(), s.currentTimestamp.String())
		return
	}
	s.currentTimestamp = updatedTimestamp
}

func (s *TimestampPersister) GetCurrentTimestamp() time.Time {
	return s.currentTimestamp
}

func getConfigMap(client *kubernetes.Clientset) (*v1.ConfigMap, error) {
	return client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), timestampConfigMapName, metav1.GetOptions{})
}
func (s *TimestampPersister) Flush() {
	s.currentTimestamp = time.Now().UTC()
	s.logger.Logger.Debug().Msgf("flushing current timstamp %s to ConfigMap", s.currentTimestamp.String())

	// stop the ticker and return from the goroutine
	s.ticker.Stop()
	s.done <- true
	s.updateConfigMap()
}
