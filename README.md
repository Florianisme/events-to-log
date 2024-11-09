## Events to Log
A simple tool which attaches to the Kube-API and listens for events.<br>
It then formats these events and prints them to the console.

### What is this trying to solve?
* Kubernetes events are lost after 1h (by default)
  * Events are not easily searchable
* I can not use my existing Log-Collection Stack (such as [ELK](https://www.elastic.co/de/elastic-stack)) on events

### Example Output
When scaling up a Deployment, this is what the output would look like:

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

### Deploying the Application
TODO Docker Image and kube-resources