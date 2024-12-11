package logging

type LoggableEvent struct {
	Metadata           `json:"event_metadata"`
	Message            string `json:"message"`
	CreationTimestamp  string `json:"timestamp"`
	FirstSeenTimestamp string `json:"firstSeenTimestamp"`
	LastSeenTimestamp  string `json:"lastSeenTimestamp"`
	Reason             string `json:"reason"`
	Type               string `json:"type"`
	Count              int32  `json:"count"`
	Reporter           string `json:"reporter"`
}

type Metadata struct {
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	UID             string `json:"uid"`
	ResourceVersion string `json:"resource_version"`
}
