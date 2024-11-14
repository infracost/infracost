package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getCodeEngineAppRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_code_engine_app",
		RFunc: newCodeEngineApp,
	}
}

func newCodeEngineApp(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	cpu := d.Get("scale_cpu_limit").String()
	memory := d.Get("scale_memory_limit").String()
	scaleinitialinstances := d.Get("scale_initial_instances").Int()
	r := &ibm.CodeEngineApp{
		Address: d.Address,
		Region:  region,
		CPU:	 cpu,
		Memory: memory,
		ScaleInitialInstances: scaleinitialinstances,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region
	configuration["cpu"] = cpu
	configuration["memory"] = memory
	configuration["scaleinitialinstances"] = scaleinitialinstances

	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
