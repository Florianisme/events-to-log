## Events to Log
A simple tool which attaches to the Kube-API and listens for events.<br>
It then formats these events and prints them to the console.

### What is this trying to solve?
* Kubernetes events are lost after 1h (by default)
  * Events are not easily searchable
* I can not use my existing Log-Collection Stack (such as [ELK](https://www.elastic.co/de/elastic-stack)) on events


## Deploying the Application
To deploy the application in your Kubernetes Cluster, run the following command:

```shell
kubctl apply -f https://raw.githubusercontent.com/Florianisme/events-to-log/refs/heads/main/resources/k8s/deployment.yaml
```

A Pod in the newly created `event-logging` namespace will then start logging all events.


### Example Output
This is what the output of another Pod's startup event would look like:

```json
{
  "count": 3,
  "creationTimestamp": "2024-12-15 12:19:48 +0100 CET",
  "eventMetadata": {
    "name": "bookinfo-gateway-istio.181155b9f9950646",
    "namespace": "default",
    "resource_version": "73535",
    "uid": "86dd8157-24f2-4790-93d6-d58d1c792031"
  },
  "firstSeenTimestamp": "2024-12-15 12:19:48 +0100 CET",
  "lastSeenTimestamp": "2024-12-15 12:46:09 +0100 CET",
  "message": "Scaled down replica set bookinfo-gateway-istio-754c86f49 to 2 from 3",
  "reason": "ScalingReplicaSet",
  "reporter": "deployment-controller",
  "type": "Normal"
}
```

## Configuration
The application is configurable via the environment variables:

| Environment Variable | Description                                                                                                                                         | Default Value |
|:--------------------:|-----------------------------------------------------------------------------------------------------------------------------------------------------|:-------------:|
|          TZ          | Timezone of the logged event's timestamp (UTC for example)                                                                                          |      UTC      |
|      NAMESPACE       | Namespace in which this application is running. If not set, tries to detect it automatically from the injected environment variable "POD_NAMESPACE" | event-logging |
|      LOG_LEVEL       | The application's internal logging level (not related to events)                                                                                    |     INFO      |