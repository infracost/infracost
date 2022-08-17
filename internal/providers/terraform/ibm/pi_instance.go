package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getPiInstanceRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "ibm_pi_instance",
		RFunc:               newPiInstance,
		ReferenceAttributes: []string{"pi_storage_type"},
	}
}

func newPiInstance(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	systemType := d.Get("pi_sys_type").String()
	processorMode := d.Get("pi_proc_type").String()
	cpus := d.Get("pi_processors").Float()
	memory := d.Get("pi_memory").Float()
	operatingSystem := d.Get("pi_image_id").String()

	refs := d.References("pi_storage_type")

	var storageType string
	for _, a := range refs {
		storageType = a.Get("pi_image_storage_type").String()
	}

	r := &ibm.PiInstance{
		Address:         d.Address,
		Region:          region,
		SystemType:      systemType,
		ProcessorMode:   processorMode,
		Cpus:            cpus,
		Memory:          memory,
		StorageType:     storageType,
		OperatingSystem: operatingSystem,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
