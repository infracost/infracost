package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudFunctionsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_cloudfunctions_function",
		RFunc: NewCloudFunctionsFunction,
	}
}

func NewCloudFunctionsFunction(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.CloudFunctionsFunction{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	if !d.IsEmpty("available_memory_mb") {
		r.AvailableMemoryMB = intPtr(d.Get("available_memory_mb").Int())
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
