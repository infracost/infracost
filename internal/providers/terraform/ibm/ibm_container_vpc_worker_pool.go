package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getIbmContainerVpcWorkerPoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_container_vpc_worker_pool",
		RFunc: newIbmContainerVpcWorkerPool,
	}
}

func newIbmContainerVpcWorkerPool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var asd = d.Get("entitlement").String()
	entitlement := false
	if asd != "" {
		entitlement = true
	}

	r := &ibm.IbmContainerVpcWorkerPool{
		Address:     d.Address,
		Region:      d.Get("region").String(),
		Flavor:      d.Get("flavor").String(),
		WorkerCount: d.Get("worker_count").Int(),
		ZoneCount:   int64(len(d.Get("zones").Array())),
		Entitlement: entitlement,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
