package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// ComputeInstance struct represents Compute Instance resource.
type ComputeInstance struct {
	Address string
	Region  string

	MachineType       string
	PurchaseOption    string
	Size              int64
	HasBootDisk       bool
	BootDiskSize      float64
	BootDiskType      string
	ScratchDisks      int
	GuestAccelerators []*ComputeGuestAccelerator
}

// ComputeInstanceUsageSchema defines a list which represents the usage schema of ComputeInstance.
var ComputeInstanceUsageSchema = []*schema.UsageItem{}

// PopulateUsage parses the u schema.UsageData into the ComputeInstance.
// It uses the `infracost_usage` struct tags to populate data into the ComputeInstance.
func (r *ComputeInstance) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid ComputeInstance struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *ComputeInstance) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		computeCostComponent(r.Region, r.MachineType, r.PurchaseOption, r.Size),
	}

	if r.HasBootDisk {
		costComponents = append(costComponents, bootDiskCostComponent(r.Region, r.BootDiskSize, r.BootDiskType))
	}

	if r.ScratchDisks > 0 {
		costComponents = append(costComponents, scratchDiskCostComponent(r.Region, r.PurchaseOption, r.ScratchDisks))
	}

	for _, guestAccel := range r.GuestAccelerators {
		costComponents = append(costComponents, guestAcceleratorCostComponent(r.Region, r.PurchaseOption, guestAccel.Type, guestAccel.Count, r.Size))
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    ComputeInstanceUsageSchema,
		CostComponents: costComponents,
	}
}
