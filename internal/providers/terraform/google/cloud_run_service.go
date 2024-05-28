package google

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudRunServiceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_cloud_run_service",
		RFunc: newCloudRunService,
	}
}

func newCloudRunService(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("location").String()
	cpuThrottling := true
	minScale := float64(0.5)
	annotations := d.Get("metadata.0.annotations").Map()
	limits := d.Get("template.0.spec.0.containers.0.resources.0.limits").Map()
	if val, ok := annotations["run.googleapis.com/cpu-throttling"]; ok {
		cpuThrottling = val.Bool()
	}
	if val, ok := annotations["autoscaling.knative.dev/minScale"]; ok {
		minScale = val.Float()
	}

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
	r := &google.CloudRunService{
		Address:             d.Address,
		Region:              region,
		CpuLimit:            cpu,
		MinInstanceCount:    minScale,
		IsThrottlingEnabled: cpuThrottling,
		MemoryLimit:         memory,
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
