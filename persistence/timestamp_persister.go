package persistence

import (
	"context"
	"events-to-log/logging"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
	"time"
)

var namespace = "event-logging"
var timestampConfigMapName = "timestamp"

type TimestampPersister struct {
	client           *kubernetes.Clientset
	ticker           *time.Ticker
	currentTimestamp time.Time
	logger           *logging.Logger
}

func Init(client *kubernetes.Clientset, logger *logging.Logger) *TimestampPersister {
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
			currentTimestamp = time.UnixMilli(int64(convertedTimestamp))
			logger.Logger.Debug().Msgf("timestamp to pick up at found, starting event logging at %s", currentTimestamp.String())
		}
	}

	ticker := time.NewTicker(5 * time.Second)

	persister := &TimestampPersister{
		client:           client,
		ticker:           ticker,
		currentTimestamp: currentTimestamp,
		logger:           logger,
	}

	go persister.updateConfigMap()
	return persister
}

func (s *TimestampPersister) updateConfigMap() {
	for range s.ticker.C {
		configMap, err := getConfigMap(s.client)
		if err != nil {
			panic(err)
		}

		configMap.Data["current"] = strconv.FormatInt(s.currentTimestamp.UnixMilli(), 10)
		s.logger.Logger.Debug().Msgf("updating current timestamp to %s (unix millis: %d",
			s.currentTimestamp.String(), s.currentTimestamp.UnixMilli())

		_, err = s.client.CoreV1().ConfigMaps(namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
		if err != nil {
			panic(err)
		}
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
	s.currentTimestamp = time.Now()
	s.logger.Logger.Debug().Msgf("flushing current timstamp %s to ConfigMap", s.currentTimestamp.String())
	s.ticker.Stop()

	s.updateConfigMap()
}
