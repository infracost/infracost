package base

type Filter struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Operation string `json:"operation,omitempty"`
}
