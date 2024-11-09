## Events to Log
A simple tool which attaches to the Kube-API and listens for events.<br>
It then formats these events and prints them to the console.

### What is this trying to solve?
* Kubernetes events are lost after 1h (by default)
  * Events are not easily searchable
* I can not use my existing Log-Collection Stack (such as [ELK](https://www.elastic.co/de/elastic-stack)) on events

### Deploying the Application
TODO Docker Image and kube-resources