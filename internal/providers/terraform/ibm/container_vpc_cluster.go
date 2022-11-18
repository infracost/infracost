package ibm

import (
	"regexp"

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

	region := d.Get("region").String()
	kubeVersion := d.Get("kube_version").String()
	flavor := d.Get("flavor").String()
	workerCount := d.Get("worker_count").Int()

	r := &ibm.ContainerVpcCluster{
		Name:        d.Get("name").String(),
		VpcId:       d.Get("vpc_id").String(),
		Region:      region,
		Flavor:      flavor,
		KubeVersion: kubeVersion,
		WorkerCount: workerCount,
		Zones:       zones,
		Entitlement: entitlement,
	}
	r.PopulateUsage(u)

	configuration := make(map[string]any)
	configuration["region"] = region
	configuration["flavor"] = flavor
	configuration["kube_version"] = kubeVersion
	configuration["worker_count"] = workerCount
	configuration["zones_count"] = len(zones)
	configuration["ocp_entitlement"] = entitlement

	resourceType := d.Type
	isRoks, _ := regexp.MatchString("(?i)openshift", kubeVersion)
	if isRoks {
		resourceType = "roks"
	}
	SetCatalogMetadata(d, resourceType, configuration)

	return r.BuildResource()
}
