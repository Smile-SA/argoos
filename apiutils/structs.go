package apiutils

type Container struct {
	Image string `json:"image"`
	Name  string `json:"name"`
}

type TSpec struct {
	Containers []Container `json:"containers"`
}

type Template struct {
	Spec TSpec `json:"spec"`
}

type Spec struct {
	Replicas int      `json:"replicas"`
	Template Template `json:"template"`
}

type MetaData struct {
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace"`
	Labels    map[string]interface{} `json:"labels"`
}

type Item struct {
	Metadata MetaData `json:"metadata"`
	Spec     Spec     `json:"spec"`
}

type APIResponse struct {
	Kind  string
	Items []Item `json:"items"`
}

type Target struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag,omitempty"`
}

type Request struct {
	Host string `json:"host"`
}

type Event struct {
	Action  string  `json:"action"`
	Target  Target  `json:"target"`
	Request Request `json:"request"`
}

type Events struct {
	Events []Event `json:"events"`
}

type Update struct {
	Containers []Container `json:"containers"`
}
