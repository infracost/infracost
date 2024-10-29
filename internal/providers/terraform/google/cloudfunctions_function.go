package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudFunctionsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_cloudfunctions_function",
		CoreRFunc: NewCloudFunctionsFunction,
	}
}

func NewCloudFunctionsFunction(d *schema.ResourceData) schema.CoreResource {
	r := &google.CloudFunctionsFunction{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	if !d.IsEmpty("available_memory_mb") {
		r.AvailableMemoryMB = intPtr(d.Get("available_memory_mb").Int())
	}

	return r
}
