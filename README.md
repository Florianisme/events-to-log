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
  "metadata": {
    "name": "bookinfo-gateway-istio.1806526e9f2c7280",
    "namespace": "default",
    "uid": "5e5e4f53-80bd-4794-8e06-7473d0dbff38"
  },
  "message": "Scaled up replica set bookinfo-gateway-istio-b485b9449 to 2 from 1",
  "timestamp": "2024-11-09 15:15:41 +0100 CET",
  "reason": "ScalingReplicaSet",
  "type": "Normal",
  "count": 1,
  "reporter": "deployment-controller"
}
```

## Configuration
The application is configurable via the environment variables:

| Environment Variable | Description                                                                                                                                         | Default Value |
|:--------------------:|-----------------------------------------------------------------------------------------------------------------------------------------------------|:-------------:|
|          TZ          | Timezone of the logged event's timestamp (UTC for example)                                                                                          |      UTC      |
|      NAMESPACE       | Namespace in which this application is running. If not set, tries to detect it automatically from the injected environment variable "POD_NAMESPACE" | event-logging |
|      LOG_LEVEL       | The application's internal logging level (not related to events)                                                                                    |     INFO      |