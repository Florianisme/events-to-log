package logging

type LoggableEvent struct {
	Metadata  `json:"metadata"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Reason    string `json:"reason"`
	Type      string `json:"type"`
	Count     int32  `json:"count"`
	Reporter  string `json:"reporter"`
}

type Metadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
}
