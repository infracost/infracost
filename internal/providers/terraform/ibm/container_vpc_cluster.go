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
	r := &ibm.ContainerVpcCluster{
		Name:        d.Get("name").String(),
		VpcId:       d.Get("vpc_id").String(),
		KubeVersion: d.Get("kube_version").String(),
		Flavor:      d.Get("flavor").String(),
		WorkerCount: d.Get("worker_count").Int(),
		Region:      d.Get("region").String(),
		ZoneCount:   int64(len(d.Get("zones").Array())),
		Entitlement: entitlement,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
