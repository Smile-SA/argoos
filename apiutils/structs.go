package apiutils

// Target is the image to change in k8s struct.
type Target struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag,omitempty"`
}

// Request contains "host" from a Docker repository event.
type Request struct {
	Host string `json:"host"`
}

// Event sent by Docker registry webhook.
type Event struct {
	Action  string  `json:"action"`
	Target  Target  `json:"target"`
	Request Request `json:"request"`
}

// Events list from Docker webhook.
type Events struct {
	Events []Event `json:"events"`
}
