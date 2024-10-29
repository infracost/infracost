package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
	"k8s.io/apimachinery/pkg/api/resource"
)

func getCloudRunV2JobRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_cloud_run_v2_job",
		CoreRFunc: newCloudRunV2Job,
	}
}

func newCloudRunV2Job(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("location").String()
	limits := d.Get("template.0.template.0.containers.0.resources.0.limits").Map()
	taskCount := int64(1)
	if !d.IsEmpty("template.0.task_count") {
		taskCount = int64(d.Get("template.0.task_count").Int())
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
	return &google.CloudRunV2Job{
		Address:     d.Address,
		Region:      region,
		CpuLimit:    cpu,
		MemoryLimit: memory,
		TaskCount:   taskCount,
	}
}
