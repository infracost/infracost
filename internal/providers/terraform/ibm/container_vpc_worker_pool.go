package ibm

import (
	"regexp"

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

	region := d.Get("region").String()
	flavor := d.Get("flavor").String()
	workerCount := d.Get("worker_count").Int()

	r := &ibm.ContainerVpcWorkerPool{
		Address:     d.Address,
		Region:      region,
		KubeVersion: kubeVersion,
		Flavor:      flavor,
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
