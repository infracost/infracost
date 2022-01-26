package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeRouterNATRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_router_nat",
		RFunc: NewComputeRouterNAT,
	}
}

func NewComputeRouterNAT(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.ComputeRouterNAT{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
