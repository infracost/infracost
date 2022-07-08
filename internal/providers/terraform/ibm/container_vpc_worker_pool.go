package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getContainerVpcWorkerPoolRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "ibm_container_vpc_worker_pool",
		RFunc:               newContainerVpcWorkerPool,
		ReferenceAttributes: []string{"cluster"},
	}
}

func newContainerVpcWorkerPool(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	refs := d.References("cluster") // Get the reference
	var kubeVersion = ""
	for _, a := range refs {
		kubeVersion = a.Get("kube_version").String()
	}
	var entitlement_str = d.Get("entitlement").String()
	entitlement := false
	if entitlement_str != "" {
		entitlement = true
	}
	zones := make([]ibm.Zone, 0)
	for _, a := range d.Get("zones").Array() {
		zones = append(zones, ibm.Zone{Name: a.Get("name").String()})
	}
	r := &ibm.ContainerVpcWorkerPool{
		Address:     d.Address,
		Region:      d.Get("region").String(),
		KubeVersion: kubeVersion,
		Flavor:      d.Get("flavor").String(),
		WorkerCount: d.Get("worker_count").Int(),
		Zones:       zones,
		Entitlement: entitlement,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
