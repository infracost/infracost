package google

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudRunV2ServiceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_cloud_run_v2_service",
		RFunc: newCloudRunV2Service,
	}
}

func newCloudRunV2Service(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("location").String()
	limits := d.Get("template.0.containers.0.resources.0.limits").Map()
	cpu := int64(1)
	if val, ok := limits["cpu"]; ok {
		cpu = int64(val.Float())
	}
	memory := int64(536870912) // 512 MiB
	if val, ok := limits["memory"]; ok {
		parseMemory, err := resource.ParseQuantity(val.String())
		if err == nil {
			memory = parseMemory.Value() // bytes
		}
	}
	isCpuIdle := true
	if !d.IsEmpty("template.0.containers.0.resources.0.cpu_idle") {
		isCpuIdle = d.Get("template.0.containers.0.resources.0.cpu_idle").Bool()
	}
	minInstanceCount := float64(0.5)
	if !d.IsEmpty("template.0.scaling.0.min_instance_count") {
		minInstanceCount = d.Get("template.0.scaling.0.min_instance_count").Float()
	}
	r := &google.CloudRunService{
		Address:             d.Address,
		Region:              region,
		CpuLimit:            cpu,
		MemoryLimit:         memory,
		IsThrottlingEnabled: isCpuIdle,
		MinInstanceCount:    minInstanceCount,
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
