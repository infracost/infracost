package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getCodeEngineFunctionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_code_engine_function",
		RFunc: newCodeEngineFunction,
	}
}

func newCodeEngineFunction(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	cpu := d.Get("scale_cpu_limit").String()
	memory := d.Get("scale_memory_limit").String()
	r := &ibm.CodeEngineFunction{
		Address: d.Address,
		Region:  region,
		CPU:	 cpu,
		Memory: memory,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region
	configuration["cpu"] = cpu
	configuration["memory"] = memory

	SetCatalogMetadata(d, d.Type, configuration)

	return r.BuildResource()
}
