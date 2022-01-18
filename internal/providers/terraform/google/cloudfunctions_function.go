package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudFunctionsRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_cloudfunctions_function",
		RFunc: NewCloudfunctionsFunction,
	}
}
func NewCloudfunctionsFunction(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.CloudfunctionsFunction{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if !d.IsEmpty("available_memory_mb") {
		r.AvailableMemoryMb = intPtr(d.Get("available_memory_mb").Int())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
