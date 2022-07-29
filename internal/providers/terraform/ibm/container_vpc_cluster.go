package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
)

func getContainerVpcClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_container_vpc_cluster",
		RFunc: newContainerVpcCluster,
	}
}

func newContainerVpcCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var entitlement_str = d.Get("entitlement").String()
	entitlement := false
	if entitlement_str != "" {
		entitlement = true
	}
	zones := make([]ibm.Zone, 0)
	for _, a := range d.Get("zones").Array() {
		zones = append(zones, ibm.Zone{Name: a.Get("name").String()})
	}
	r := &ibm.ContainerVpcCluster{
		Name:        d.Get("name").String(),
		VpcId:       d.Get("vpc_id").String(),
		KubeVersion: d.Get("kube_version").String(),
		Flavor:      d.Get("flavor").String(),
		WorkerCount: d.Get("worker_count").Int(),
		Region:      d.Get("region").String(),
		Zones:       zones,
		Entitlement: entitlement,
	}
	r.PopulateUsage(u)
	SetCatalogMetadata(d, d.Type)

	return r.BuildResource()
}
