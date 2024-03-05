package google

import (
	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ContainerNodePool struct represents Container Cluster's Node Pool resource.
type ContainerNodePool struct {
	Address string
	Region  string

	Zones        int64
	CountPerZone int64
	NodeConfig   *ContainerNodeConfig

	// "usage" args
	Nodes *int64 `infracost_usage:"nodes"`
}

func (r *ContainerNodePool) CoreType() string {
	return "ContainerNodePool"
}

func (r *ContainerNodePool) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "nodes", DefaultValue: 0, ValueType: schema.Int64},
	}
}

// PopulateUsage parses the u schema.UsageData into the ContainerNodePool.
// It uses the `infracost_usage` struct tags to populate data into the ContainerNodePool.
func (r *ContainerNodePool) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ContainerNodePool struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ContainerNodePool) BuildResource() *schema.Resource {
	countPerZone := r.CountPerZone
	if r.Nodes != nil {
		countPerZone = *r.Nodes
	}

	nodeCount := decimal.NewFromInt(r.Zones * countPerZone)

	poolSize := int64(1)

	costComponents, err := computeCostComponents(r.Region, r.NodeConfig.MachineType, r.NodeConfig.PurchaseOption, poolSize, nil)
	if err != nil {
		logging.Logger.Warn().Msgf("Skipping resource %s. %s", r.Address, err)
		return nil
	}

	costComponents = append(costComponents, computeDiskCostComponent(r.Region, r.NodeConfig.DiskType, r.NodeConfig.DiskSize, poolSize))

	if r.NodeConfig.LocalSSDCount > 0 {
		costComponents = append(costComponents, scratchDiskCostComponent(r.Region, r.NodeConfig.PurchaseOption, int(r.NodeConfig.LocalSSDCount)))
	}

	for _, guestAccel := range r.NodeConfig.GuestAccelerators {
		if component := guestAcceleratorCostComponent(r.Region, r.NodeConfig.PurchaseOption, guestAccel.Type, guestAccel.Count, poolSize, nil); component != nil {
			costComponents = append(costComponents, component)
		}
	}

	resource := &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}

	schema.MultiplyQuantities(resource, nodeCount)

	return resource
}
