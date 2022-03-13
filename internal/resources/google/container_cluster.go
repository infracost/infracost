package google

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// ContainerCluster struct represents Container Cluster resource.
type ContainerCluster struct {
	Address string
	Region  string

	IsZone          bool
	DefaultNodePool *ContainerNodePool
	NodePools       []*ContainerNodePool

	// "usage" args
	DefaultNodePoolNodes *int64 `infracost_usage:"nodes"`
}

// ContainerClusterUsageSchema defines a list which represents the usage schema of ContainerCluster.
// Nested wildcard node_pool usage is mapped on provider level.
var ContainerClusterUsageSchema = []*schema.UsageItem{
	{Key: "nodes", DefaultValue: 0, ValueType: schema.Int64},
}

// PopulateUsage parses the u schema.UsageData into the ContainerCluster.
// It uses the `infracost_usage` struct tags to populate data into the ContainerCluster.
func (r *ContainerCluster) PopulateUsage(u *schema.UsageData) {
	if u == nil {
		return
	}

	resources.PopulateArgsWithUsage(r, u)

	if r.DefaultNodePool != nil {
		r.DefaultNodePool.PopulateUsage(u)
	}

	for i, nodePool := range r.NodePools {
		nodePool.PopulateUsage(&schema.UsageData{
			Attributes: u.Get(fmt.Sprintf("node_pool[%d]", i)).Map(),
		})
	}
}

// BuildResource builds a schema.Resource from a valid ContainerCluster struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ContainerCluster) BuildResource() *schema.Resource {
	description := "Regional Kubernetes Clusters"
	if r.IsZone {
		description = "Zonal Kubernetes Clusters"
	}

	costComponents := []*schema.CostComponent{
		{
			Name:           "Cluster management fee",
			Unit:           "hours",
			UnitMultiplier: decimal.NewFromInt(1),
			HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("gcp"),
				Region:        strPtr("global"),
				Service:       strPtr("Kubernetes Engine"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "description", Value: strPtr(description)},
				},
			},
			PriceFilter: &schema.PriceFilter{
				StartUsageAmount: strPtr("0"),
				EndUsageAmount:   strPtr(""),
			},
		},
	}

	subresources := []*schema.Resource{}

	if r.DefaultNodePool != nil {
		r.DefaultNodePool.Address = "default_pool"
		subresources = append(subresources, r.DefaultNodePool.BuildResource())
	}

	for i, nodePool := range r.NodePools {
		nodePool.Address = fmt.Sprintf("node_pool[%d]", i)
		subresources = append(subresources, nodePool.BuildResource())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    ContainerClusterUsageSchema,
		CostComponents: costComponents,
		SubResources:   subresources,
	}
}
