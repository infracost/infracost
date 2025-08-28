package google

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ContainerCluster struct represents Container Cluster resource.
type ContainerCluster struct {
	Address string
	Region  string

	AutopilotEnabled bool

	IsZone          bool
	DefaultNodePool *ContainerNodePool
	NodePools       []*ContainerNodePool

	// "usage" args
	DefaultNodePoolNodes        *int64   `infracost_usage:"nodes"`
	AutopilotVCPUCount          *float64 `infracost_usage:"autopilot_vcpu_count"`
	AutopilotMemoryGB           *float64 `infracost_usage:"autopilot_memory_gb"`
	AutopilotEphemeralStorageGB *float64 `infracost_usage:"autopilot_ephemeral_storage_gb"`
}

func (r *ContainerCluster) CoreType() string {
	return "ContainerCluster"
}

func (r *ContainerCluster) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "nodes", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "autopilot_vcpu_count", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "autopilot_memory_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "autopilot_ephemeral_storage_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
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

	for _, nodePool := range r.NodePools {
		nodePool.PopulateUsage(&schema.UsageData{
			Attributes: u.Get(nodePool.Address).Map(),
		})
	}
}

// BuildResource builds a schema.Resource from a valid ContainerCluster struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ContainerCluster) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, r.managementFeeCostComponent())

	if r.AutopilotEnabled {
		costComponents = append(costComponents, r.autopilotCPUCostComponent())
		costComponents = append(costComponents, r.autopilotMemoryCostComponent())
		costComponents = append(costComponents, r.autopilotStorageCostComponent())
	}

	subresources := []*schema.Resource{}

	if r.DefaultNodePool != nil {
		poolResource := r.DefaultNodePool.BuildResource()
		if poolResource != nil {
			subresources = append(subresources, poolResource)
		}
	}

	for _, nodePool := range r.NodePools {
		poolResource := nodePool.BuildResource()
		if poolResource != nil {
			subresources = append(subresources, poolResource)
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
		SubResources:   subresources,
	}
}

// managementFeeCostComponent returns a cost component for cluster management
// fee.
func (r *ContainerCluster) managementFeeCostComponent() *schema.CostComponent {
	description := "Regional Kubernetes Clusters"
	name := "Cluster management fee"

	if r.IsZone {
		description = "Zonal Kubernetes Clusters"
	}

	if r.AutopilotEnabled {
		description = "Autopilot Kubernetes Clusters"
		name = "Autopilot"
	}

	return &schema.CostComponent{
		Name:           name,
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
	}
}

// autopilotCPUCostComponent returns a cost component for Autopilot vCPU usage.
func (r *ContainerCluster) autopilotCPUCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	multiplier := decimal.NewFromInt(1000) // Price is for mCPU

	if r.AutopilotVCPUCount != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.AutopilotVCPUCount).Mul(multiplier))
	}

	return &schema.CostComponent{
		Name:           "Autopilot vCPU",
		Unit:           "vCPU",
		UnitMultiplier: schema.HourToMonthUnitMultiplier.Mul(multiplier),
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Kubernetes Engine"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr("^Autopilot Pod CPU Requests")},
			},
		},
		UsageBased: true,
	}
}

// autopilotMemoryCostComponent returns a cost component for Autopilot memory usage.
func (r *ContainerCluster) autopilotMemoryCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.AutopilotMemoryGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.AutopilotMemoryGB))
	}

	return &schema.CostComponent{
		Name:           "Autopilot memory",
		Unit:           "GB",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Kubernetes Engine"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr("^Autopilot Pod Memory Requests")},
			},
		},
		UsageBased: true,
	}
}

// autopilotStorageCostComponent returns a cost component for Autopilot
// ephemeral storage usage.
func (r *ContainerCluster) autopilotStorageCostComponent() *schema.CostComponent {
	var quantity *decimal.Decimal
	if r.AutopilotEphemeralStorageGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.AutopilotEphemeralStorageGB))
	}

	return &schema.CostComponent{
		Name:           "Autopilot ephemeral storage",
		Unit:           "GB",
		UnitMultiplier: schema.HourToMonthUnitMultiplier,
		HourlyQuantity: quantity,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Kubernetes Engine"),
			ProductFamily: strPtr("Compute"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", ValueRegex: regexPtr("^Autopilot Pod Ephemeral Storage Requests")},
			},
		},
		UsageBased: true,
	}
}
