package crossplane

import (
	"strings"

	"github.com/tidwall/gjson"
)

func getProvider(apiVersion string) string {
	var provider string
	data := strings.Split(apiVersion, "/")
	if len(data) > 0 {
		provider = data[0]
	}
	return provider
}

func getLabels(parsed gjson.Result) map[string]string {
	labels := make(map[string]string)
	for k, v := range parsed.Get("metadata.labels").Map() {
		labels[k] = v.String()
	}
	return labels
}
