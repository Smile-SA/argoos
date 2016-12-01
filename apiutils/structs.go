package apiutils

// Container represents a container in "spec.containers" K8s struct.
type Container struct {
	Image string `json:"image"`
	Name  string `json:"name"`
}

// TSpec contains "containers" from "template.spec" k8s struct.
type TSpec struct {
	Containers []Container `json:"containers"`
}

// Template is a spec.template representation from k8s struct.
type Template struct {
	Spec TSpec `json:"spec"`
}

// Spec is a .spec representation from k8s struct.
type Spec struct {
	Replicas int      `json:"replicas"`
	Template Template `json:"template"`
}

// Metadata from a k8s struct.
type MetaData struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Labels    map[string]interface{} `json:"labels"`
}

// Item is a k8s list item from k8s response.
type Item struct {
	Metadata MetaData `json:"metadata"`
	Spec     Spec     `json:"spec"`
}

// APIResponse represent a k8s List struct.
type APIResponse struct {
	Kind  string
	Items []Item `json:"items"`
}

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

// Update is a container list to be patched in k8s.
type Update struct {
	Containers []Container `json:"containers"`
}
